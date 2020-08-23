from rest_framework import serializers

from . import models


class HostSerializer(serializers.HyperlinkedModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.Host
        fields = "__all__"


class ProcessSerializer(serializers.HyperlinkedModelSerializer):
    id = serializers.ReadOnlyField()
    machine_id = serializers.CharField()

    class Meta:
        model = models.Process
        fields = "__all__"
        extra_kwargs = {
            "machine_id": {"write_only": True}
        }


class ProcessEventSerializer(serializers.HyperlinkedModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.ProcessEvent
        fields = "__all__"


class ProcessConfigurationSerializer(serializers.HyperlinkedModelSerializer):
    id = serializers.ReadOnlyField()

    class Meta:
        model = models.ProcessConfiguration
        fields = "__all__"
