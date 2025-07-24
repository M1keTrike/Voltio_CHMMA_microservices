# 🚀 GUÍA RÁPIDA - Deployment Voltio en AWS EC2 Ubuntu

## 📋 Resumen
He creado un sistema completo que empaqueta todos tus consumers y el WebSocket server en un único contenedor Docker con Supervisor para gestionar todos los procesos.

## 🏗️ Archivos Creados

### Principales:
- `Dockerfile` - Construye todos los servicios en un contenedor
- `docker-compose.yml` - Orquesta el deployment
- `.env.example` - Template de configuración
- `deploy-aws.sh` - Script de deployment automático para Ubuntu
- `start-voltio.ps1` - Script para desarrollo en Windows

### Documentación:
- `README-DEPLOYMENT.md` - Guía completa de deployment
- `health-check.sh` - Script de verificación de servicios

## 🚀 INSTRUCCIONES PASO A PASO

### 1. Preparar AWS EC2

1. **Crear instancia EC2** Ubuntu 20.04/22.04
2. **Security Group**: Permitir puertos 22 (SSH) y 8081 (WebSocket)
3. **Conectar por SSH**:
   ```bash
   ssh -i "tu-clave.pem" ubuntu@tu-ec2-ip
   ```

### 2. Deployment Automático

```bash
# En la instancia EC2, ejecutar:
curl -fsSL https://raw.githubusercontent.com/M1keTrike/Voltio_CHMMA/main/deploy-aws.sh -o deploy-aws.sh
chmod +x deploy-aws.sh
./deploy-aws.sh
```

### 3. Configurar Variables

```bash
# Editar configuración
nano /opt/voltio/Voltio_CHMMA/.env

# Configurar al menos:
RABBITMQ_URI=amqp://admin:trike@tu-rabbitmq:5672/
INFLUX_URL=http://tu-influxdb:8086
INFLUX_TOKEN=tu-token
```

### 4. Iniciar Servicios

```bash
# Aplicar grupo docker
newgrp docker

# Ir al directorio
cd /opt/voltio/Voltio_CHMMA

# Iniciar
sudo systemctl start voltio-services

# Verificar
docker-compose ps
```

## 🌐 Acceso

- **WebSocket Server**: `http://tu-ec2-ip:8081`
- **Health Check**: `http://tu-ec2-ip:8081/health`

## 📊 Monitoreo

```bash
# Ver logs
docker-compose logs -f

# Estado de servicios
docker-compose exec voltio-services supervisorctl status

# Verificar salud
./health-check.sh
```

## 🔧 Comandos Útiles

```bash
# Parar servicios
sudo systemctl stop voltio-services

# Reiniciar servicios
sudo systemctl restart voltio-services

# Ver logs del sistema
journalctl -u voltio-services -f

# Reconstruir después de cambios
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 🏠 Desarrollo Local (Windows)

```powershell
# En PowerShell, desde el directorio del proyecto:
.\start-voltio.ps1          # Iniciar
.\start-voltio.ps1 -Logs    # Ver logs
.\start-voltio.ps1 -Stop    # Parar
```

## 🆘 Troubleshooting Rápido

1. **Servicios no inician**: `docker-compose logs`
2. **Error de conectividad**: Verificar Security Group y .env
3. **Puerto no accesible**: `sudo ufw status` y verificar 8081
4. **Health check falla**: `./health-check.sh`

## 📞 Verificación Final

Después del deployment, verificar:

1. ✅ `docker-compose ps` muestra "healthy"
2. ✅ `curl http://localhost:8081/health` responde OK
3. ✅ `./health-check.sh` muestra todo verde
4. ✅ Logs sin errores: `docker-compose logs`

¡Tu sistema Voltio está listo para producción! 🎉
