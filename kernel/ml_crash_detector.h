#ifndef MODULE_H
#define MODULE_H

#include <linux/ftrace.h>
#include <linux/types.h>
#include <linux/kallsyms.h>
#include <linux/kernel.h>
#include <linux/linkage.h>
#include <linux/module.h>
#include <linux/slab.h>
#include <linux/uaccess.h>
#include <linux/version.h>
#include <linux/sched.h>
#include <linux/time.h>
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
#include <linux/netlink.h>
#include <linux/fs.h>
#include <asm/siginfo.h>
#include <asm/io.h>
#include <asm/pgalloc.h>
#include <asm/tlb.h>
#include <asm/tlbflush.h>
#include <asm/pgtable.h>
#include <asm/uaccess.h>
#include <asm/atomic.h>
#include <net/netlink.h>
#include <net/net_namespace.h>

/*
 * Note: Minimum support Kernel version is 3.17 (determined by use of linux/rhashtable.h).
 */

MODULE_DESCRIPTION("Memlab's crash detector");

MODULE_AUTHOR("Memlab - Tal Hoffman");

MODULE_LICENSE("GPL"); // todo: not GPL?

#define NETLINK_MEMLAB_FAMILY 25
#define NETLINK_GROUP_MONITOR_PROCESS 1
#define NETLINK_GROUP_SIGNALS 2

#ifndef CONFIG_X86_64
#error Only supporting x86_64 architecture
#endif

#if defined(CONFIG_X86_64) && (LINUX_VERSION_CODE >= KERNEL_VERSION(4, 17, 0))
#define NEW_SYSCALL_NAMING_CONVENTION 1
#endif

#ifdef NEW_SYSCALL_NAMING_CONVENTION
#define SYSCALL_NAME(name) ("__x64_" name)
#else
#define SYSCALL_NAME(name) (name)
#endif

#define SIGNAL_MASK(sig) (1 << (sig - 1))

const int wait_timeout_seconds = 60;

struct ftrace_hook {
    const char *name;
    void *new_func;
    void *orig_func;
    unsigned long orig_address;
    struct ftrace_ops ops;
};

struct task_struct *get_task_by_pid(pid_t pid);

// todo: decide final mask
static u32 relevant_signals_mask =
        SIGNAL_MASK(SIGINT) |
        SIGNAL_MASK(SIGHUP) |
        SIGNAL_MASK(SIGQUIT) |
        SIGNAL_MASK(SIGILL) |
        SIGNAL_MASK(SIGTRAP) |
        SIGNAL_MASK(SIGABRT) |
        SIGNAL_MASK(SIGIOT) |
        //SIGNAL_MASK(SIGBUS)    |
        //SIGNAL_MASK(SIGFPE)    |
        SIGNAL_MASK(SIGKILL) |
        SIGNAL_MASK(SIGUSR1) |
        SIGNAL_MASK(SIGUSR2) |
        SIGNAL_MASK(SIGSEGV)   |
        //SIGNAL_MASK(SIGPIPE)   |
        SIGNAL_MASK(SIGALRM) |
        SIGNAL_MASK(SIGTERM) |
        SIGNAL_MASK(SIGSTKFLT) |
        //SIGNAL_MASK(SIGCHLD)   |
        SIGNAL_MASK(SIGCONT) |
        SIGNAL_MASK(SIGSTOP) |
        SIGNAL_MASK(SIGTSTP) |
        SIGNAL_MASK(SIGTTIN) |
        SIGNAL_MASK(SIGTTOU) |
        SIGNAL_MASK(SIGURG) |
        SIGNAL_MASK(SIGXCPU) |
        SIGNAL_MASK(SIGXFSZ) |
        SIGNAL_MASK(SIGVTALRM) |
        SIGNAL_MASK(SIGPROF) |
        SIGNAL_MASK(SIGWINCH) |
        SIGNAL_MASK(SIGIO) |
        SIGNAL_MASK(SIGPOLL);

bool is_signal_relevant(int sig);

bool is_task_relevant(struct task_struct *dst);

static int resolve_hooked_func_address(struct ftrace_hook *hook);

static void notrace

ftrace_hook_callback(unsigned long ip, unsigned long parent_ip,
                     struct ftrace_ops *ops, struct pt_regs *regs);

int setup_hook(struct ftrace_hook *hook);

void teardown_hook(struct ftrace_hook *hook);

int setup_hooks(struct ftrace_hook *hooks, size_t count);

void teardown_hooks(struct ftrace_hook *hooks, size_t count);

static asmlinkage void internal_kill(pid_t pid, int sig);

static asmlinkage void internal_force_sig(int sig, struct task_struct *to);

#endif