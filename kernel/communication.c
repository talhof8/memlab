#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/version.h>
#include <linux/printk.h>
#include <linux/skbuff.h>
#include <net/genetlink.h>
#include "state.h"
#include "communication.h"

int setup_communication_sockets() {
    int error;

#if LINUX_VERSION_CODE >= KERNEL_VERSION(4, 10, 0)
    error = genl_register_family(&ml_usr_to_kern_family);
    if (error < 0) {
        return error;
    }

    return genl_register_family(&ml_kern_to_usr_family);
#else
    error = genl_register_family_with_ops_groups(&ml_usr_to_kern_family, usr_to_kern_ops, usr_to_kern_groups);
    if (error < 0) {
        return error;
    }

    return genl_register_family_with_ops_groups(&ml_kern_to_usr_family, kern_to_usr_ops, kern_to_usr_groups);
#endif
}

void release_communication_sockets() {
    int error = genl_unregister_family(&ml_usr_to_kern_family);
    if (error < 0) {
        pr_err("[ML Crash Detector] Failed to unregister family (utk): %d.\n", error);
        // Best effort - do not return
    }

    error = genl_unregister_family(&ml_kern_to_usr_family);
    if (error < 0) {
        pr_err("[ML Crash Detector] Failed to unregister family (ktu): %d.\n", error);
        // Best effort - do not return
    }
}

static int utk_handle_monitor_process_command(struct sk_buff *skb, struct genl_info *info) {
    pr_info("[ML Crash Detector] utk_handle_monitor_process_command() start.\n");

    if (!info->attrs[ML_ATTRIBUTE_PID]) {
        pr_err("[ML Crash Detector] Invalid request: missing 'pid' attribute.\n");
        return -EINVAL;
    } else if (!info->attrs[ML_ATTRIBUTE_MONITOR_DO_WATCH]) {
        pr_err("[ML Crash Detector] Invalid request: missing 'watch' attribute.\n");
        return -EINVAL;
    }

    pid_t pid = (pid_t) nla_get_u32(info->attrs[ML_ATTRIBUTE_PID]);
    uint8_t do_watch = nla_get_u8(info->attrs[ML_ATTRIBUTE_MONITOR_DO_WATCH]);

    int err;
    if (do_watch) {
        pr_info("[ML Crash Detector] Watch process '%d'.\n", pid);

        if (is_process_watched(pid)) {
            pr_info("[ML Crash Detector] Process is already watched: %d.\n", pid);
            return -EINVAL;
        }

        if (!add_process_to_watched_processes(pid)) {
            pr_err("[ML Crash Detector] Failed to add process to watched processes: (pid: %d, err: %d).\n", pid, err);
            return -EINVAL;
        }
    } else {
        pr_info("[ML Crash Detector] Unwatch process '%d'.\n", pid);

        if (!is_process_watched(pid)) {
            pr_info("[ML Crash Detector] Process is not watched: %d.\n", pid);
            return -EINVAL;
        }

        if (!remove_process_from_watched_processes(pid)) {
            pr_err("[ML Crash Detector] Failed to remove process from watched processes: (pid: %d, err: %d).\n", pid,
                   err);
            return -EINVAL;
        }
    }

    pr_info("[ML Crash Detector] utk_handle_monitor_process_command() done.\n");
    return 0;
}

static int utk_handle_handled_caught_signal_command(struct sk_buff *skb, struct genl_info *info) {
    pr_info("[ML Crash Detector] utk_handle_handled_caught_signal_command() start.\n");

    if (!info->attrs[ML_ATTRIBUTE_PID]) {
        pr_err("[ML Crash Detector] Invalid request: missing 'pid' attribute.\n");
        return -EINVAL;
    }

    pid_t pid = (pid_t) nla_get_u32(info->attrs[ML_ATTRIBUTE_PID]);

    interrupt_watched_process_wait_queue(pid);

    pr_info("[ML Crash Detector] utk_handle_handled_caught_signal_command() done.\n");
    return 0;
}

static int ktu_handle_notify_caught_signal_command(struct sk_buff *skb, struct genl_info *info) {
    return 0;
}

int ktu_send_caught_signal_notification(pid_t pid, uint32_t signal) {
    pr_info("[ML Crash Detector] ktu_send_caught_signal_notification(%d, %d, %s) start.\n", pid, signal);

    struct sk_buff *skb;
    void *msg_head;
    int success, err = 0;

    skb = genlmsg_new(NLMSG_GOODSIZE, GFP_KERNEL);
    if (!skb) {
        pr_err("[ML Crash Detector] genlmsg_new() failed.\n");
        goto exit;
    }

    msg_head = genlmsg_put(skb, 0, 0, &ml_kern_to_usr_family, 0, ML_COMMAND_NOTIFY_CAUGHT_SIGNAL);
    if (!msg_head) {
        pr_err("[ML Crash Detector] genlmsg_put() failed.\n");
        kfree_skb(skb);
        goto exit;
    }

    err = nla_put_u32(skb, ML_ATTRIBUTE_PID, pid);
    if (err) {
        pr_err("[ML Crash Detector] Pid nla_put_u32() failed: %d.\n", err);
        kfree_skb(skb);
        goto exit;
    }

    err = nla_put_u32(skb, ML_ATTRIBUTE_SIGNAL_NOTIFICATION_SIGNAL, signal);
    if (err) {
        pr_err("[ML Crash Detector] Signal nla_put_u32() failed: %d.\n", err);
        kfree_skb(skb);
        goto exit;
    }

    genlmsg_end(skb, msg_head);

    err = genlmsg_multicast_allns(&ml_kern_to_usr_family, skb, 0, 0, GFP_KERNEL);
    if (err) {
        pr_warn("genlmsg_multicast_allns() failed (perhaps no one is listening): %d.\n", err);
        goto exit;
    }

    success = 1;

    exit:
    pr_info("[ML Crash Detector] Send caught-signal notification successfully (pid: %d, signal: %d).\n",
            pid, signal);
    return success;
}