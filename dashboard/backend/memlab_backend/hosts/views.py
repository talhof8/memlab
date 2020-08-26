from datetime import timedelta

from django.utils import timezone
from memlab_backend.hosts import models, serializers
from rest_framework import viewsets, mixins, status, decorators
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response


class HostViewSet(mixins.ListModelMixin, mixins.RetrieveModelMixin, mixins.CreateModelMixin, mixins.UpdateModelMixin,
                  viewsets.GenericViewSet):
    queryset = models.Host.objects.all()
    serializer_class = serializers.HostSerializer
    lookup_field = "machine_id"

    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data, many=True)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data

        try:
            machine_id = validated_data["machine_id"]
        except KeyError as e:
            raise ValidationError(e)

        # todo: support sending a partial list and omit empty values

        host = models.Host.objects.update_or_create(machine_id=machine_id, defaults=validated_data)

        data = serializer.to_representation(host)
        return Response(data, status=status.HTTP_201_CREATED)

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
        serializer = self.get_serializer(data=request.data, many=True)
        serializer.is_valid(raise_exception=True)

        try:
            for item in serializer.validated_data:
                _ = models.Process.objects.update_or_create(pid=item["pid"], machine_id=item["machine_id"],
                                                            defaults=item)
        except KeyError as e:
            raise ValidationError(e)

        return Response(status=status.HTTP_201_CREATED)

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


class DetectionConfigViewSet(viewsets.ModelViewSet):
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