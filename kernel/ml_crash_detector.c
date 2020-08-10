#include <linux/ftrace.h>
#include <linux/kallsyms.h>
#include <linux/kernel.h>
#include <linux/printk.h>
#include <linux/linkage.h>
#include <linux/module.h>
#include <linux/slab.h>
#include <linux/version.h>
#include <linux/sched.h>
#include <linux/sched/task.h>
#include <linux/kprobes.h>
#include <linux/mm.h>
#include <linux/stacktrace.h>
#include <linux/hugetlb.h>
#include <linux/mman.h>
#include <linux/swap.h>
#include <linux/highmem.h>
#include <linux/pagemap.h>
#include <linux/ksm.h>
#include <linux/rmap.h>
#include <linux/delayacct.h>
#include <linux/init.h>
#include <linux/writeback.h>
#include <linux/memcontrol.h>
#include <linux/mmu_notifier.h>
#include <linux/elf.h>
#include <linux/delay.h>
#include <linux/fs.h>
#include <linux/uaccess.h>
#include <asm/siginfo.h>
#include <asm/io.h>
#include <asm/pgalloc.h>
#include <asm/tlb.h>
#include <asm/tlbflush.h>
#include <asm/pgtable.h>
#include <asm/uaccess.h>
#include <asm/atomic.h>
#include "ml_crash_detector.h"
#include "state.h"
#include "communication.h"

// todo: do not allow to add ml's agent pid

// todo: move to some process.h (and also impl inside state.h)
struct task_struct *get_task_by_pid(pid_t pid) {
    struct task_struct *t;
    struct pid *p;

    rcu_read_lock();
    p = find_vpid(pid);
    if (!p) {
        rcu_read_unlock();
        return NULL;
    }

    t = pid_task(p, PIDTYPE_PID);
    get_task_struct(t);
    rcu_read_unlock();
    return t;
}

bool is_signal_relevant(int sig) {
    if (sig > SIGSYS || sig <= 0) { // SIGSYS is 31.
        return false;
    }

    if (relevant_signals_mask & SIGNAL_MASK(sig)) {
        return true;
    }

    return false;
}

bool is_task_relevant(struct task_struct *dst) {
    if (!dst || dst->pid == 0) {
        return false;
    }

    return is_process_watched(dst->pid);
}

static int resolve_hooked_func_address(struct ftrace_hook *hook) {
    hook->orig_address = kallsyms_lookup_name(hook->name);

    if (!hook->orig_address) {
        pr_info("[ML Crash Detector] unresolved symbol: %s\n", hook->name);
        return -ENOENT;
    }

    // MCOUNT_INSN_SIZE - sizeof mcount call (see ftrace.h).
    // We need to skip passed it.
    *((unsigned long *) hook->orig_func) = hook->orig_address + MCOUNT_INSN_SIZE;

    return 0;
}

static void notrace

ftrace_hook_callback(unsigned long ip, unsigned long parent_ip,
                     struct ftrace_ops *ops, struct pt_regs *regs) {
    struct ftrace_hook *hook = container_of(ops,
                                            struct ftrace_hook, ops);
    regs->ip = (unsigned long) hook->new_func;
}

int setup_hook(struct ftrace_hook *hook) {
    int err;

    err = resolve_hooked_func_address(hook);
    if (err) {
        return err;
    }

    hook->ops.func = ftrace_hook_callback;
    hook->ops.flags = FTRACE_OPS_FL_SAVE_REGS | FTRACE_OPS_FL_RECURSION_SAFE | FTRACE_OPS_FL_IPMODIFY;

    err = ftrace_set_filter_ip(&hook->ops, hook->orig_address, 0, 0);
    if (err) {
        pr_info("[ML Crash Detector] ftrace_set_filter_ip() failed: %d.\n", err);
        return err;
    }

    err = register_ftrace_function(&hook->ops);
    if (err) {
        pr_info("[ML Crash Detector] register_ftrace_function() failed: %d.\n", err);
        ftrace_set_filter_ip(&hook->ops, hook->orig_address, 1, 0);
        return err;
    }

    return 0;
}

int setup_hooks(struct ftrace_hook *hooks, size_t count) {
    int err;
    size_t i;

    for (i = 0; i < count; i++) {
        err = setup_hook(&hooks[i]);
        if (err) {
            goto error;
        }
    }

    return 0;

    error:
    while (i != 0) {
        teardown_hook(&hooks[--i]);
    }

    return -err;
}

void teardown_hook(struct ftrace_hook *hook) {
    int err;

    err = unregister_ftrace_function(&hook->ops);
    if (err) {
        pr_info("[ML Crash Detector] unregister_ftrace_function() failed: %d.\n", err);
    }

    err = ftrace_set_filter_ip(&hook->ops, hook->orig_address, 1, 0);
    if (err) {
        pr_info("[ML Crash Detector] ftrace_set_filter_ip() failed: %d.\n", err);
    }
}

void teardown_hooks(struct ftrace_hook *hooks, size_t count) {
    size_t i;

    for (i = 0; i < count; i++) {
        teardown_hook(&hooks[i]);
    }
}

static asmlinkage void internal_kill(pid_t pid, int sig) {
    struct task_struct *from, *to;
    int err;

    if (!is_signal_relevant(sig)) {
        return;
    }

    from = current;
    to = get_task_by_pid(pid); // Note: calls get_task_struct()

    if (!to) {
        return;
    }

    if (!is_task_relevant(to)) {
        put_task_struct(to);
        return;
    }

    pr_info(
            "===========sys_kill==========\n"
            "user:%d process:%d[%s] send SIG %d to %d[%s]\n",
            (int) from_kuid(&init_user_ns, current_uid()), from->pid, from->comm, sig, to->pid, to->comm);
    put_task_struct(to);

    if ((err = ktu_send_caught_signal_notification(pid, sig)) < 0) {
        pr_err("[ML Crash Detector] Failed to send caught-signal notification: (pid: %d, err: %d).\n", pid, err);
        return;
    }

    wait_on_watched_process_wait_queue(pid, wait_timeout_seconds);
    remove_process_from_watched_processes(pid);

    pr_info("[ML Crash Detector] ===========sys_kill==========\n");
    return;
}

#ifdef NEW_SYSCALL_NAMING_CONVENTION

static asmlinkage void (*real_sys_kill)(struct pt_regs *regs);

static asmlinkage void ml_sys_kill(struct pt_regs *regs) {
    pid_t pid;
    int sig;

    pid = regs->di;
    sig = regs->si;

    internal_kill(pid, sig);
    real_sys_kill(regs);
}

#else

static asmlinkage void (*real_sys_kill)(pid_t pid, int sig);

static asmlinkage void ml_sys_kill(pid_t pid, int sig) {
    internal_kill(pid, sig);
    real_sys_kill(pid, sig);
}

#endif

static asmlinkage void internal_prepare_signal(int sig, struct task_struct *to) {
    struct task_struct *from;
    int err;
    pid_t pid;


    if (!to) {
        goto exit;
    }

    rcu_read_lock();
    get_task_struct(to);
    rcu_read_unlock();

    pid = to->pid;

    if (!is_signal_relevant(sig)) {
        goto exit;
    }

    from = current;

    if (!is_task_relevant(to)) {
        put_task_struct(to);
        goto exit;
    }

    pr_info(
            "===========prepare_signal==========\n"
            "user:%d process:%d[%s] send SIG %d to %d[%s]\n",
            (int) from_kuid(&init_user_ns, current_uid()), from->pid, from->comm, sig, to->pid, to->comm);
    put_task_struct(to);

    if ((err = ktu_send_caught_signal_notification(pid, sig)) < 0) {
        pr_err("[ML Crash Detector] Failed to send caught-signal notification: (pid: %d, err: %d).\n", pid, err);
        goto exit;
    }

    wait_on_watched_process_wait_queue(pid, wait_timeout_seconds);
    remove_process_from_watched_processes(pid);

    pr_info("[ML Crash Detector] ===========prepare_signal==========\n");

    exit:
    return;
}


#if LINUX_VERSION_CODE > KERNEL_VERSION(3, 3, 8)

static asmlinkage void (*real_prepare_signal)(int sig, struct task_struct *p, bool force);

static asmlinkage void ml_prepare_signal(int sig, struct task_struct *p, bool force) {
    internal_prepare_signal(sig, p);
    real_prepare_signal(sig, p, force);
}

#else

static asmlinkage void (*real_prepare_signal)(int sig, struct task_struct *p, int from_ancestor_ns);

static asmlinkage void ml_prepare_signal(int sig, struct task_struct *p, int from_ancestor_ns) {
    internal_prepare_signal(sig, p);
    real_prepare_signal(sig, p, from_ancestor_ns);
}
#endif

// todo: hook crashes (__send_signal?)
static struct ftrace_hook hook_list[] = {
//        {
//                .name = (SYSCALL_NAME("sys_kill")),
//                .new_func = (ml_sys_kill),
//                .orig_func = (&real_sys_kill),
//        },
        {
                .name = ("prepare_signal"),
                .new_func = (ml_prepare_signal),
                .orig_func = (&real_prepare_signal),
        },
};

static int ml_init(void) {
    int err;

    pr_info("[ML Crash Detector] Initialize state.\n");
    if ((err = init_state()) < 0) {
        pr_err("[ML Crash Detector] Failed to initialize state: %d.\n", err);
        return err;
    }

    pr_info("[ML Crash Detector] Setup nl socket.\n");
    if ((err = setup_communication_sockets()) < 0) {
        pr_err("[ML Crash Detector] Failed to setup nl socket: %d.\n", err);
        return err;
    }

    pr_info("[ML Crash Detector] Install hooks.\n");
    if ((err = setup_hooks(hook_list, ARRAY_SIZE(hook_list))) < 0) {
        pr_err("[ML Crash Detector] Failed to install hooks: %d.\n", err);
        return err;
    }

    pr_info("[ML Crash Detector] Module loaded.\n");
    return 0;
}

static void ml_exit(void) {
    int err;

    pr_info("[ML Crash Detector] Remove hooks.\n");
    teardown_hooks(hook_list, ARRAY_SIZE(hook_list));

    pr_info("[ML Crash Detector] Release nl socket.\n");
    release_communication_sockets();

    pr_info("[ML Crash Detector] Clear state.\n");
    if ((err = clear_state()) < 0) {
        pr_err("[ML Crash Detector] Failed to clear state: %d.\n", err);
        // Not returning - best effort...
    }

    pr_info("[ML Crash Detector] Module unloaded.\n");
}

module_init(ml_init);
module_exit(ml_exit);
