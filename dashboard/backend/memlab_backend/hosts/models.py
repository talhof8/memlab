import uuid

from django.db import models
from memlab_backend.accounts import models as account_models


class Host(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    user = models.OneToOneField(account_models.User, on_delete=models.CASCADE)
    machine_id = models.CharField(max_length=150, blank=False, null=False)
    public_ip_address = models.GenericIPAddressField(null=False, blank=False)
    hostname = models.CharField(max_length=500, blank=True, null=True)
    first_seen = models.DateTimeField(auto_now_add=True)
    last_probe_at = models.DateTimeField(auto_now=True)
    last_boot_at = models.DateTimeField(null=True, blank=True)
    operating_system = models.CharField(max_length=20, null=True, blank=True)  # E.g: freebsd, linux
    platform = models.CharField(max_length=150, null=True, blank=True)  # E.g: ubuntu, linuxmint
    platform_family = models.CharField(max_length=150, null=True, blank=True)  # E.g: debian, rhel
    platform_version = models.CharField(max_length=150, null=True, blank=True)
    kernel_version = models.CharField(max_length=20, null=True, blank=True)
    kernel_architecture = models.CharField(max_length=20, null=True, blank=True)
    virtualization_system = models.CharField(max_length=150, null=True, blank=True)
    virtualization_role = models.CharField(max_length=10, null=True, blank=True)  # Guest or host


class Process(models.Model):
    STATUS_RUNNING = 'R'
    STATUS_SLEEP = 'S'
    STATUS_STOP = 'T'
    STATUS_IDLE = 'I'
    STATUS_ZOMBIE = 'Z'
    STATUS_WAIT = 'W'
    STATUS_LOCK = 'L'
    STATUSES = [
        (STATUS_RUNNING, "Running"),
        (STATUS_SLEEP, "Sleep"),
        (STATUS_STOP, "Stop"),
        (STATUS_IDLE, "Idle"),
        (STATUS_ZOMBIE, "Zombie"),
        (STATUS_WAIT, "Wait"),
        (STATUS_LOCK, "Lock"),
    ]

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    user = models.OneToOneField(account_models.User, on_delete=models.CASCADE)
    host = models.ForeignKey(Host, on_delete=models.CASCADE, null=False, blank=False)
    pid = models.IntegerField(null=False, blank=False)
    executable = models.CharField(max_length=255, null=False, blank=False)
    command_line = models.CharField(max_length=1000, null=False, blank=False)
    create_time = models.DateTimeField(null=False, blank=False)
    last_seen_at = models.DateTimeField(auto_now=True)
    monitored = models.BooleanField(default=False)
    monitored_since = models.DateTimeField(null=True, blank=True)
    status = models.CharField(max_length=1, choices=STATUSES, default=STATUS_RUNNING)

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
    user = models.OneToOneField(account_models.User, on_delete=models.CASCADE)
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


class DetectionConfig(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    user = models.OneToOneField(account_models.User, on_delete=models.CASCADE)
    process = models.ForeignKey(Process, on_delete=models.CASCADE, null=False, blank=False)
    created_at = models.DateTimeField(auto_now_add=True)
    modified_at = models.DateTimeField(auto_now=True)
    detect_signals = models.BooleanField(default=False)
    detect_thresholds = models.BooleanField(default=False)
    detect_suspected_hangs = models.BooleanField(default=False)
    cpu_threshold = models.IntegerField(null=True, blank=True)
    memory_threshold = models.IntegerField(null=True, blank=True)
    suspected_hang_duration = models.DurationField(null=True, blank=True)
    restart_on_signal = models.BooleanField(default=True)
    restart_on_cpu_threshold = models.BooleanField(default=False)
    restart_on_memory_threshold = models.BooleanField(default=False)
    restart_on_suspected_hang = models.BooleanField(default=False)
