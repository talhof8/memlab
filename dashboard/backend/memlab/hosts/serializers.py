from rest_framework import serializers

from . import models


class HostSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = models.Host
        fields = "__all__"


class ProcessSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = models.Process
        fields = "__all__"


class ProcessEventSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = models.ProcessEvent
        fields = "__all__"


class ProcessConfigurationSerializer(serializers.HyperlinkedModelSerializer):
    class Meta:
        model = models.ProcessConfiguration
        fields = "__all__"
