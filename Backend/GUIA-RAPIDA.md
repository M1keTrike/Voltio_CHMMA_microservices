# 📋 Guía Rápida - Backend Voltio

## 🚀 Para empezar YA (3 pasos)

### 1️⃣ Subir con FileZilla
- Sube la carpeta `Backend/` completa a tu servidor
- No olvides subir en modo **binario** para los archivos `.sh`

### 2️⃣ Dar permisos
```bash
chmod +x voltio-manager.sh
```

### 3️⃣ ¡A funcionar!
```bash
# Todos los consumers juntos (RECOMENDADO)
./voltio-manager.sh all

# O uno por uno
./voltio-manager.sh single pzem
./voltio-manager.sh single dht22
```

## 🔍 Comandos Útiles

### Ver qué está pasando
```bash
./voltio-manager.sh status    # Estado de contenedores
./voltio-manager.sh logs      # Logs de todos
./voltio-manager.sh logs pzem # Logs de PZEM solamente
```

### Controlar servicios
```bash
./voltio-manager.sh stop      # Parar todo
./voltio-manager.sh restart   # Reiniciar todo
./voltio-manager.sh help      # Ver todas las opciones
```

## 🔧 ¿Qué hace cada consumer?

| Consumer | Para qué sirve | Base de datos |
|----------|---------------|--------------|
| `pzem` | 🔌 Datos eléctricos (voltaje, corriente, potencia) | InfluxDB |
| `dht22` | 🌡️ Temperatura y humedad | InfluxDB |
| `pir` | 👁️ Detección de movimiento | InfluxDB |
| `light` | 💡 Niveles de luz | InfluxDB |
| `notification` | 📧 Alertas y notificaciones | PostgreSQL |
| `automation` | 🤖 Reglas automáticas | PostgreSQL |

## 🆘 Si algo falla

### Ver errores
```bash
./voltio-manager.sh logs | grep -i error
```

### Reiniciar un service específico
```bash
./voltio-manager.sh stop
./voltio-manager.sh single pzem  # Solo el que falló
```

### Verificar conexiones
- **RabbitMQ**: Puerto 5672 (admin: 15672)
- **InfluxDB**: Puerto 8086
- **PostgreSQL**: 13.222.89.227:5432

## 💡 Tips

- **Logs en tiempo real**: `./voltio-manager.sh logs -f`
- **Solo ver un consumer**: `./voltio-manager.sh logs pzem`
- **Para desarrollo**: Usa consumers individuales
- **Para producción**: Usa `./voltio-manager.sh all`

## 🎯 Casos de uso comunes

### Desarrollo/Testing
```bash
# Probar solo datos eléctricos
./voltio-manager.sh single pzem

# Probar sensores ambientales
./voltio-manager.sh single dht22
./voltio-manager.sh single light
```

### Producción completa
```bash
# Todo funcionando
./voltio-manager.sh all

# Monitorear
./voltio-manager.sh status
./voltio-manager.sh logs
```

---
**¿Dudas?** Lee el `README-DOCKER.md` completo para más detalles.
