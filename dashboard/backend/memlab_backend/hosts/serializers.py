from memlab_backend.hosts import models
from rest_framework import serializers


class HostSerializer(serializers.ModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.Host
        fields = "__all__"
        read_only_fields = ["id", "user"]


class ProcessSerializer(serializers.ModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.Process
        fields = "__all__"
        read_only_fields = ["id", "user", "host"]


class ProcessCreateSerializer(serializers.Serializer):
    id = serializers.ReadOnlyField()
    machine_id = serializers.CharField(max_length=models.Host.MACHINE_ID_LENGTH,
                                       min_length=models.Host.MACHINE_ID_LENGTH)
    processes = ProcessSerializer(many=True)

    @classmethod
    def many_init(cls, *args, **kwargs):
        raise NotImplementedError()

    def create(self, validated_data):
        return NotImplementedError()

    def update(self, instance, validated_data):
        return NotImplementedError()

    class Meta:
        fields = "__all__"
        extra_kwargs = {
            "machine_id": {"write_only": True}
        }


class ProcessEventSerializer(serializers.ModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.ProcessEvent
        fields = "__all__"
        read_only_fields = ["id", "user"]


class DetectionConfigSerializer(serializers.ModelSerializer):
    id = serializers.ReadOnlyField()
    pid = serializers.SerializerMethodField("get_pid")
    machine_id = serializers.SerializerMethodField("get_machine_id")
    process_create_time = serializers.SerializerMethodField("get_process_create_time")

    class Meta:
        model = models.DetectionConfig
        read_only_fields = ["id"]
        exclude = ["user", "process"]

    def get_pid(self, obj):
        return obj.process.pid

    def get_machine_id(self, obj):
        return obj.process.host.machine_id

    def get_process_create_time(self, obj):
        return obj.process.create_time
