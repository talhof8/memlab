# Generated by Django 3.1 on 2020-08-25 13:00

from django.conf import settings
from django.db import migrations, models
import django.db.models.deletion
import django.utils.timezone
import uuid


class Migration(migrations.Migration):

    dependencies = [
        migrations.swappable_dependency(settings.AUTH_USER_MODEL),
        ('hosts', '0011_host_machine_id'),
    ]

    operations = [
        migrations.CreateModel(
            name='DetectionConfig',
            fields=[
                ('id', models.UUIDField(default=uuid.uuid4, editable=False, primary_key=True, serialize=False)),
                ('created_at', models.DateTimeField(auto_now_add=True)),
                ('modified_at', models.DateTimeField(auto_now=True)),
                ('detect_signals', models.BooleanField(default=False)),
                ('detect_thresholds', models.BooleanField(default=False)),
                ('detect_suspected_hangs', models.BooleanField(default=False)),
                ('cpu_threshold', models.IntegerField(blank=True, null=True)),
                ('memory_threshold', models.IntegerField(blank=True, null=True)),
                ('suspected_hang_duration', models.DurationField(blank=True, null=True)),
                ('restart_on_signal', models.BooleanField(default=True)),
                ('restart_on_cpu_threshold', models.BooleanField(default=False)),
                ('restart_on_memory_threshold', models.BooleanField(default=False)),
                ('restart_on_suspected_hang', models.BooleanField(default=False)),
            ],
        ),
        migrations.RemoveField(
            model_name='host',
            name='last_keepalive_at',
        ),
        migrations.RemoveField(
            model_name='process',
            name='disappeared_at',
        ),
        migrations.RemoveField(
            model_name='process',
            name='is_active',
        ),
        migrations.RemoveField(
            model_name='process',
            name='seen_at',
        ),
        migrations.AddField(
            model_name='host',
            name='last_probe_at',
            field=models.DateTimeField(auto_now=True),
        ),
        migrations.AddField(
            model_name='host',
            name='user',
            field=models.OneToOneField(default='560fa6ea-a912-440b-9228-b379d564a273', on_delete=django.db.models.deletion.CASCADE, to='accounts.user'),
            preserve_default=False,
        ),
        migrations.AddField(
            model_name='process',
            name='create_time',
            field=models.DateTimeField(default=django.utils.timezone.now),
            preserve_default=False,
        ),
        migrations.AddField(
            model_name='process',
            name='last_seen_at',
            field=models.DateTimeField(auto_now=True),
        ),
        migrations.AddField(
            model_name='process',
            name='status',
            field=models.CharField(choices=[('R', 'Running'), ('S', 'Sleep'), ('T', 'Stop'), ('I', 'Idle'), ('Z', 'Zombie'), ('W', 'Wait'), ('L', 'Lock')], default='R', max_length=1),
        ),
        migrations.AddField(
            model_name='process',
            name='user',
            field=models.OneToOneField(default='560fa6ea-a912-440b-9228-b379d564a273', on_delete=django.db.models.deletion.CASCADE, to='accounts.user'),
            preserve_default=False,
        ),
        migrations.AddField(
            model_name='processevent',
            name='user',
            field=models.OneToOneField(default='560fa6ea-a912-440b-9228-b379d564a273', on_delete=django.db.models.deletion.CASCADE, to='accounts.user'),
            preserve_default=False,
        ),
        migrations.DeleteModel(
            name='ProcessConfiguration',
        ),
        migrations.AddField(
            model_name='detectionconfig',
            name='process',
            field=models.ForeignKey(on_delete=django.db.models.deletion.CASCADE, to='hosts.process'),
        ),
        migrations.AddField(
            model_name='detectionconfig',
            name='user',
            field=models.OneToOneField(on_delete=django.db.models.deletion.CASCADE, to=settings.AUTH_USER_MODEL),
        ),
    ]