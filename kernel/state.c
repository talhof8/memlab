#include <linux/rcu_sync.h>
#include <linux/memory.h>
#include <linux/memcontrol.h>
#include <linux/sched/task.h>
#include <linux/wait.h>
#include <linux/jiffies.h>
#include <linux/refcount.h>
#include "state.h"

int init_state() {
    return rhashtable_init(&watched_processes, &hashable_obj_params);
}

int clear_state() {
    rhashtable_destroy(&watched_processes); // also frees objects.
    return 0;
}

static bool process_exists_in_os(pid_t pid) {
    struct pid *vpid;

    rcu_read_lock();
    vpid = find_vpid(pid);
    bool exists = (vpid != NULL);
    rcu_read_unlock();

    return exists;
}

static struct watched_process_t *get_watched_process_unsafe(pid_t pid) {
    return rhashtable_lookup_fast(&watched_processes, &pid, hashable_obj_params);
}

static bool is_process_watched_unsafe(pid_t pid) {
    struct watched_process_t *watched_process;
    watched_process = get_watched_process_unsafe(pid);
    return (watched_process != NULL);
}

bool is_process_watched(pid_t pid) {
    spin_lock(&state_lock);
    int success = is_process_watched_unsafe(pid);
    spin_unlock(&state_lock);

    return success;
}

void wait_on_watched_process_wait_queue(pid_t pid, unsigned long timeout_secs) {
    pr_info("[ML Crash Detector] wait_on_watched_process_wait_queue(%d) start.\n", pid);

    struct watched_process_t *watched_process;

    if (!is_process_watched_unsafe(pid)) {
        pr_info("[ML Crash Detector] wait_on_watched_process_wait_queue(%d) done (pid not watched).\n", pid);
        return;
    }

    spin_lock(&state_lock);

    if (!is_process_watched_unsafe(pid)) { // Double-checked locking.
        pr_info("[ML Crash Detector] wait_on_watched_process_wait_queue(%d) done (pid not watched).\n", pid);
        spin_unlock(&state_lock);
        return;
    }

    watched_process = get_watched_process_unsafe(pid);
    acquire_watched_process(watched_process);
    spin_unlock(&state_lock);

    pr_info("[ML Crash Detector] Waiting until handled (timeout: %d seconds)", timeout_secs);
    wait_event_interruptible_timeout(watched_process->wait_queue,
                                     atomic_read(&watched_process->handled) == 1,
                                     timeout_secs * HZ);

    // Intentionally freeing here so we do not block the spin lock while waiting for the event.
    release_watched_process(watched_process);

    pr_info("[ML Crash Detector] wait_on_watched_process_wait_queue(%d) done.\n", pid);
}

void interrupt_watched_process_wait_queue(pid_t pid) {
    pr_info("[ML Crash Detector] interrupt_watched_process_wait_queue(%d) start.\n", pid);

    struct watched_process_t *watched_process;

    if (!is_process_watched_unsafe(pid)) {
        return;
    }

    spin_lock(&state_lock);

    if (!is_process_watched_unsafe(pid)) { // Double-checked locking.
        spin_unlock(&state_lock);
        return;
    }

    watched_process = get_watched_process_unsafe(pid);
    acquire_watched_process(watched_process);

    atomic_set(&watched_process->handled, 1);
    wake_up_interruptible(&watched_process->wait_queue);

    release_watched_process(watched_process);
    spin_unlock(&state_lock);

    pr_info("[ML Crash Detector] interrupt_watched_process_wait_queue(%d) done.\n", pid);
}

bool add_process_to_watched_processes(pid_t pid) {
    pr_info("[ML Crash Detector] add_process_to_watched_processes(%d) start.\n", pid);

    int success = false;
    spin_lock(&state_lock);

    if (is_process_watched_unsafe(pid)) {
        pr_info("[ML Crash Detector] Process is already watched: %d.\n", pid);
        goto exit;
    } else if (!process_exists_in_os(pid)) {
        pr_err("[ML Crash Detector] Process cannot be found in os: %d.\n", pid);
        goto exit;
    }

    struct watched_process_t *watched_process;
    watched_process = (struct watched_process_t *) kmalloc(sizeof(struct watched_process_t), GFP_KERNEL);
    watched_process->pid = pid;
    atomic_set(&watched_process->handled, 0);
    init_waitqueue_head(&watched_process->wait_queue);
    refcount_set(&watched_process->ref, 0);

    // Should be finally released when removing process from watched processes list.
    acquire_watched_process(watched_process);

    int err = rhashtable_insert_fast(&watched_processes, &watched_process->linkage, hashable_obj_params);
    if (err != 0) {
        release_watched_process(watched_process);
        pr_err("[ML Crash Detector] rhashtable_lookup_insert_fast() failed: %d.\n", err);
        goto exit;
    }

    success = true;

    exit:
    spin_unlock(&state_lock);
    pr_info("[ML Crash Detector] add_process_to_watched_processes(%d) done.\n", pid);
    return success;
}

bool remove_process_from_watched_processes(pid_t pid) {
    pr_info("[ML Crash Detector] remove_process_from_watched_processes(%d) start.\n", pid);

    int success = false;
    spin_lock(&state_lock);

    struct watched_process_t *watched_process;
    watched_process = get_watched_process_unsafe(pid);
    if (!watched_process) {
        goto exit;
    }

    rhashtable_remove_fast(&watched_processes, &watched_process->linkage, hashable_obj_params);

    // Should be first acquired when inserting process to watched processes list.
    release_watched_process(watched_process);

    success = true;

    exit:
    spin_unlock(&state_lock);

    pr_info("[ML Crash Detector] remove_process_from_watched_processes(%d) done.\n", pid);
    return success;
}

static void acquire_watched_process(struct watched_process_t *watched_process) {
    refcount_inc(&watched_process->ref);
}

static void release_watched_process(struct watched_process_t *watched_process) {
    if (refcount_dec_and_test(&watched_process->ref)) {
        int pid = watched_process->pid;
        kfree(watched_process);
        pr_info("[ML Crash Detector] Freed watched process '%d'.\n", pid);
    }
}