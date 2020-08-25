from django.urls import include, path
from memlab_backend.accounts import views
from rest_framework import routers

router = routers.DefaultRouter()
router.register(r'company', views.CompanyViewSet, basename='company')
router.register(r'user', views.UserViewSet, basename='user')
router.register(r'', views.ProfileViewSet, basename='profile')  # Internal view will mount it to /profile

urlpatterns = [
    path('', include(router.urls)),
]
