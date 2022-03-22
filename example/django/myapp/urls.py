from django.urls import path

from . import views

urlpatterns = [
    path("", views.IndexView.as_view(), name="index"),
    path("hello/<str:username>", views.HelloView.as_view(), name="hello"),
    path("failing", views.FailingView.as_view(), name="failing"),
]
