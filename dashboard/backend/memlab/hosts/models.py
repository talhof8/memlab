import uuid

from django.db import models
from django.utils import timezone


class Host(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    ip_address = models.GenericIPAddressField(null=True, blank=True)
    hostname = models.CharField(max_length=500, blank=True, null=True)
    last_activity = models.CharField(max_length=150, null=False, blank=False)
    last_activity_at = models.DateTimeField(null=False, blank=False)


class Process(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    host = models.ForeignKey(Host, on_delete=models.CASCADE, null=False, blank=False)
    pid = models.IntegerField(null=False, blank=False)
    executable = models.CharField(max_length=255, null=False, blank=False)
    command_line = models.CharField(max_length=1000, null=False, blank=False)
    is_active = models.BooleanField(default=True)
    monitored = models.BooleanField(default=False)
    seen_at = models.DateTimeField(default=timezone.now, null=False, blank=False)
    monitored_since = models.DateTimeField(null=True, blank=True)
    disappeared_at = models.DateTimeField(null=True, blank=True)

    @classmethod
    def get_monitored_processes_by_host_ip(cls, host_ip):
        return cls.objects.filter(host__ip_address=host_ip)


class ProcessEvent(models.Model):
    TYPE_SEEN = 'A'
    TYPE_CAUGHT_SIGNAL = 'B'
    TYPE_CPU_THRESHOLD_REACHED = 'C'
    TYPE_MEMORY_THRESHOLD_REACHED = 'D'
    TYPE_SUSPECTED_HANG_CAUGHT = 'E'
    TYPE_EXITED = 'F'
    TYPE_NOT_FOUND = 'G'
    TYPES = [
        (TYPE_SEEN, "Seen"),
        (TYPE_CAUGHT_SIGNAL, "Signal caught"),
        (TYPE_CPU_THRESHOLD_REACHED, "CPU threshold reached"),
        (TYPE_MEMORY_THRESHOLD_REACHED, "Memory threshold reached"),
        (TYPE_SUSPECTED_HANG_CAUGHT, "Suspected hang caught"),
        (TYPE_EXITED, "Exited"),
        (TYPE_NOT_FOUND, "Not found"),
    ]

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    process = models.ForeignKey(Process, on_delete=models.CASCADE, null=False, blank=False)
    type = models.CharField(max_length=1, choices=TYPES, default=TYPE_SEEN)
    created_at = models.DateTimeField(auto_now_add=True)
    caught_signal = models.IntegerField(null=True, blank=True)
    cpu_usage = models.IntegerField(null=True, blank=True)
    memory_usage = models.IntegerField(null=True, blank=True)
    exit_code = models.IntegerField(null=True, blank=True)
    core_dump_location = models.URLField(null=True, blank=True)

    @classmethod
    def get_all_events(cls, process):
        return cls.objects.filter(process=process)

    @classmethod
    def get_latest_event(cls, process):
        return cls.objects.filter(process=process).latest('created_at')


class ProcessConfiguration(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    process = models.ForeignKey(Process, on_delete=models.CASCADE, null=False, blank=False)
    monitor_signals = models.BooleanField(default=False)
    monitor_thresholds = models.BooleanField(default=False)
    monitor_suspected_hangs = models.BooleanField(default=False)
    cpu_threshold = models.IntegerField(null=True, blank=True)
    memory_threshold = models.IntegerField(null=True, blank=True)
    suspected_hang_duration = models.DurationField(null=True, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    modified_at = models.DateTimeField(auto_now=True)
