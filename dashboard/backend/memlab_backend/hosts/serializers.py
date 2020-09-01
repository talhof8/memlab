from memlab_backend.hosts import models
from rest_framework import serializers


class HostSerializer(serializers.ModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.Host
        fields = "__all__"
        read_only_fields = ["id", "user"]


class ProcessSerializer(serializers.HyperlinkedModelSerializer):
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


class ProcessEventSerializer(serializers.HyperlinkedModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.ProcessEvent
        fields = "__all__"
        read_only_fields = ["id", "user"]


class DetectionConfigSerializer(serializers.HyperlinkedModelSerializer):
    id = serializers.ReadOnlyField()
    pid = serializers.SerializerMethodField("get_pid")
    machine_id = serializers.SerializerMethodField("get_machine_id")

    class Meta:
        model = models.DetectionConfig
        fields = "__all__"
        read_only_fields = ["id", "user"]

    def get_pid(self, obj):
        return obj.process.pid

    def get_machine_id(self, obj):
        return obj.process.machine_id
