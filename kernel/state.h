#ifndef STATE_H
#define STATE_H

#include <linux/spinlock.h>
#include <linux/rhashtable.h>
#include <linux/wait.h>
#include <linux/refcount.h>

struct watched_process_t {
    pid_t pid;
    atomic_t handled;
    wait_queue_head_t wait_queue;
    refcount_t ref;
    struct rhash_head linkage;
};

const static struct rhashtable_params hashable_obj_params = {
        .key_len     = sizeof(int),
        .key_offset  = offsetof(struct watched_process_t, pid),
        .head_offset = offsetof(struct watched_process_t, linkage),
};

static struct rhashtable watched_processes;

static DEFINE_SPINLOCK(state_lock);

int init_state(void);

int clear_state(void);

static bool process_exists_in_os(pid_t pid);

static struct watched_process_t *get_watched_process_unsafe(pid_t pid);

static bool is_process_watched_unsafe(pid_t pid);

bool is_process_watched(pid_t pid);

void wait_on_watched_process_wait_queue(pid_t pid, unsigned long timeout);

void interrupt_watched_process_wait_queue(pid_t pid);

bool add_process_to_watched_processes(pid_t pid);

bool remove_process_from_watched_processes(pid_t pid);

static void acquire_watched_process(struct watched_process_t *watched_process);

static void release_watched_process(struct watched_process_t *watched_process);

#endif