#ifndef COMMUNICATION_H
#define COMMUNICATION_H

#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/version.h>
#include <net/genetlink.h>
#include <linux/printk.h>

#define MEMLAB_USR_TO_KERN_FAMILY "memlab-utk"
#define MEMLAB_KERN_TO_USR_FAMILY "memlab-ktu"

enum ml_usr_to_kern_commands {
    // Do no change order of anything, this is ABI!
    ML_COMMAND_MONITOR_PROCESS,
    ML_COMMAND_HANDLED_CAUGHT_SIGNAL,
};

enum ml_kern_to_usr_commands {
    // Do no change order of anything, this is ABI!
    ML_COMMAND_NOTIFY_CAUGHT_SIGNAL,
};

enum ml_detector_attribute_ids {
    // Do no change order of anything, this is ABI!
    // Numbering must start from 1
    ML_ATTRIBUTE_PID = 1,
    ML_ATTRIBUTE_MONITOR_DO_WATCH,
    ML_ATTRIBUTE_SIGNAL_NOTIFICATION_SIGNAL,

    // This is a special one, don't list any more after this.
    ML_ATTRIBUTE_COUNT,
#define ML_ATTRIBUTE_MAX (ML_ATTRIBUTE_COUNT - 1)
};

struct nla_policy const static ml_generic_nl_policy[ML_ATTRIBUTE_COUNT] = {
        [ML_ATTRIBUTE_PID] = {.type = NLA_U32},
        [ML_ATTRIBUTE_MONITOR_DO_WATCH] = {.type = NLA_U8}, // 1 - watch, 0 - unwatch
        [ML_ATTRIBUTE_SIGNAL_NOTIFICATION_SIGNAL] = {.type = NLA_U32},
};

int setup_communication_sockets(void);

void release_communication_sockets(void);

static int utk_handle_monitor_process_command(struct sk_buff *skb, struct genl_info *info);

static int utk_handle_handled_caught_signal_command(struct sk_buff *skb, struct genl_info *info);

static int ktu_handle_notify_caught_signal_command(struct sk_buff *skb, struct genl_info *info);

int ktu_send_caught_signal_notification(pid_t pid, uint32_t signal);

const static struct genl_ops usr_to_kern_ops[] = {
        {
                .cmd = ML_COMMAND_MONITOR_PROCESS,
                .doit = utk_handle_monitor_process_command,
#if LINUX_VERSION_CODE < KERNEL_VERSION(5, 2, 0)
                /* Before kernel 5.2, each op had its own policy. */
                .policy = ml_generic_nl_policy,
#endif
        },
        {
                .cmd = ML_COMMAND_HANDLED_CAUGHT_SIGNAL,
                .doit = utk_handle_handled_caught_signal_command,
#if LINUX_VERSION_CODE < KERNEL_VERSION(5, 2, 0)
                /* Before kernel 5.2, each op had its own policy. */
                .policy = ml_generic_nl_policy,
#endif
        },
};

const static struct genl_ops kern_to_usr_ops[] = {
        {
                .cmd = ML_COMMAND_NOTIFY_CAUGHT_SIGNAL,
                .doit = ktu_handle_notify_caught_signal_command,
#if LINUX_VERSION_CODE < KERNEL_VERSION(5, 2, 0)
                /* Before kernel 5.2, each op had its own policy. */
                .policy = ml_generic_nl_policy,
#endif
        },
};

static struct genl_multicast_group usr_to_kern_groups[] = {
        { .name = "MemlabUtkGroup" },
};

static struct genl_multicast_group kern_to_usr_groups[] = {
        { .name = "MemlabKtuGroup" },
};

static struct genl_family ml_usr_to_kern_family = {
        .name = MEMLAB_USR_TO_KERN_FAMILY,
        .version = 1,
        .maxattr = ML_ATTRIBUTE_MAX,
        .module = THIS_MODULE,
        .ops = usr_to_kern_ops,
        .n_ops = ARRAY_SIZE(usr_to_kern_ops),
#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 2, 0)
        /* Since kernel 5.2, the policy is family-wide. */
        .policy = ml_generic_nl_policy,
#endif
#if LINUX_VERSION_CODE >= KERNEL_VERSION(4, 10, 0)
        /* Since kernel 4.10.0, the groups are family-wide. */
        .mcgrps = usr_to_kern_groups,
        .n_mcgrps = ARRAY_SIZE(usr_to_kern_groups),
#endif
};

static struct genl_family ml_kern_to_usr_family = {
        .name = MEMLAB_KERN_TO_USR_FAMILY,
        .version = 1,
        .maxattr = ML_ATTRIBUTE_MAX,
        .module = THIS_MODULE,
        .ops = kern_to_usr_ops,
        .n_ops = ARRAY_SIZE(kern_to_usr_ops),
#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 2, 0)
        /* Since kernel 5.2, the policy is family-wide. */
        .policy = ml_generic_nl_policy,
#endif
#if LINUX_VERSION_CODE >= KERNEL_VERSION(4, 10, 0)
        /* Since kernel 4.10.0, the groups are family-wide. */
        .mcgrps = kern_to_usr_groups,
        .n_mcgrps = ARRAY_SIZE(kern_to_usr_groups),
#endif
};

#endif