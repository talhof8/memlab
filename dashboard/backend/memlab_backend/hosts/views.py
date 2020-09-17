from datetime import timedelta

from django.utils import timezone
from memlab_backend.hosts import models, serializers
from memlab_backend.utils.models import add_user_to_validated_data
from rest_framework import viewsets, mixins, status, decorators
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response


class HostViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.CreateModelMixin, mixins.UpdateModelMixin,
                  viewsets.GenericViewSet):
    queryset = models.Host.objects.all()
    serializer_class = serializers.HostSerializer
    lookup_field = "machine_id"

    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data, many=False)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data

        try:
            machine_id = validated_data["machine_id"]
        except KeyError as e:
            raise ValidationError(e)

        # todo: support sending a partial list and omit empty values

        add_user_to_validated_data(request, validated_data)
        host, _ = models.Host.objects.update_or_create(user__id=self.request.user.id, machine_id=machine_id,
                                                       defaults=validated_data)

        data = serializer.to_representation(host)
        return Response(data, status=status.HTTP_200_OK)

    def get_serializer_context(self):
        return {'request': None}


class ProcessViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.UpdateModelMixin, mixins.CreateModelMixin,
                     viewsets.GenericViewSet):
    queryset = models.Process.objects.all()
    serializer_class = serializers.ProcessSerializer

    @decorators.action(detail=False, methods=['get'], url_path='by_machine')
    def by_machine(self, request, machine_id):
        instances = models.Process.objects.filter(user__id=self.request.user.id, host__machine_id=machine_id)
        serializer = self.get_serializer(instances, many=True)
        return Response(serializer.data)

    def create(self, request, *args, **kwargs):
        kwargs['context'] = self.get_serializer_context()

        serializer = serializers.ProcessCreateSerializer(data=request.data, **kwargs)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data
        try:
            host = models.Host.objects.get(user__id=self.request.user.id, machine_id=validated_data["machine_id"])
        except models.Host.DoesNotExist:
            return Response(status=status.HTTP_404_NOT_FOUND)

        try:
            for item in validated_data["processes"]:
                add_user_to_validated_data(request, item)
                item["host"] = host
                _ = models.Process.objects.update_or_create(user__id=self.request.user.id, host__id=host.id,
                                                            pid=item["pid"], defaults=item)
        except KeyError as e:
            raise ValidationError(e)

        return Response(status=status.HTTP_200_OK)

    def get_serializer_context(self):
        return {'request': None}


class ProcessEventViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.CreateModelMixin,
                          viewsets.GenericViewSet):
    queryset = models.ProcessEvent.objects.all()
    serializer_class = serializers.ProcessEventSerializer

    @decorators.action(detail=False, methods=['get'], url_path='by_machine')
    def by_machine(self, request, machine_id):
        instances = models.ProcessEvent.objects.filter(user__id=self.request.user.id,
                                                       process__host__machine_id=machine_id)
        serializer = self.get_serializer(instances, many=True)
        return Response(serializer.data)

    def get_serializer_context(self):
        return {'request': None}


class DetectionConfigViewSet(mixins.CreateModelMixin, mixins.RetrieveModelMixin, mixins.UpdateModelMixin,
                             mixins.ListModelMixin, viewsets.GenericViewSet):
    # Note: intentionally not allowing /DELETE requests, because our agent won't identify deleted detection configs.

    queryset = models.DetectionConfig.objects.all()
    serializer_class = serializers.DetectionConfigSerializer

    @decorators.action(detail=False, methods=['get'], url_path='by_machine')
    def by_machine(self, request, machine_id):
        instances = models.DetectionConfig.objects.filter(user__id=self.request.user.id,
                                                          process__host__machine_id=machine_id,
                                                          process__last_seen_at__gte=self.__last_day())
        serializer = self.get_serializer(instances, many=True)
        return Response(serializer.data)

    def list(self, request, *args, **kwargs):
        instances = models.DetectionConfig.objects.filter(user__id=self.request.user.id,
                                                          process__last_seen_at__gte=self.__last_day())
        serializer = self.get_serializer(instances, many=True)
        return Response(serializer.data)

    @staticmethod
    def __last_day():
        return timezone.now().date() - timedelta(days=1)

    def get_serializer_context(self):
        return {'request': None}
