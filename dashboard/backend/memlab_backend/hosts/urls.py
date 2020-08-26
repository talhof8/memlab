from django.urls import include, path
from memlab_backend.hosts import views
from rest_framework import routers

router = routers.DefaultRouter()
router.register(r'hosts', views.HostViewSet, basename='host')
router.register(r'processes', views.ProcessViewSet, basename='process')
router.register(r'process_events', views.ProcessEventViewSet, basename='process_event')
router.register(r'detection_configs', views.DetectionConfigViewSet, basename='detection_config')

urlpatterns = [
    path('', include(router.urls)),
]