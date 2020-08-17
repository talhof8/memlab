from django.urls import include, path
from rest_framework import routers

from . import views

router = routers.DefaultRouter()
router.register(r'companies', views.CompanyViewSet, basename="company")
router.register(r'users', views.UserViewSet, basename="user")
router.register(r'profiles', views.ProfileViewSet, basename="profile")

urlpatterns = [
    path('', include(router.urls)),
]