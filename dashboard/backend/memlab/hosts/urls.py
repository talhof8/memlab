from django.urls import include, path
from rest_framework import routers

from . import views

router = routers.DefaultRouter()
router.register(r'hosts', views.HostViewSet, basename='host')
router.register(r'processes', views.ProcessViewSet, basename='process')
router.register(r'process_events', views.ProcessEventViewSet, basename='process_event')
router.register(r'process_configurations', views.ProcessConfigurationViewSet, basename='process_configuration')

urlpatterns = [
    path('', include(router.urls)),
]
