# Script para iniciar todos los productores y consumidores de Voltio
# Ejecutar desde el directorio Voltio_CHMMA

# Iniciar Productores (en nuevas ventanas)
Start-Process powershell -ArgumentList 'cd Backend/test_producers; go run dht22_producer.go'
Start-Process powershell -ArgumentList 'cd Backend/test_producers; go run pzem/main.go'
Start-Process powershell -ArgumentList 'cd Backend/test_producers; go run pir/main.go'
Start-Process powershell -ArgumentList 'cd Backend/test_producers; go run light/main.go'

# Iniciar Consumidores (en nuevas ventanas)
Start-Process powershell -ArgumentList 'cd Backend/PZEM_ConsumerSender; go run middleware/RabbitToSocketMiddleware_NEW.go'
Start-Process powershell -ArgumentList 'cd Backend/DHT22_ConsumerSender; go run middleware/RabbitToSocketMiddleware.go'
Start-Process powershell -ArgumentList 'cd Backend/PIR_ConsumerSender; go run middleware/RabbitToSocketMiddleware.go'
Start-Process powershell -ArgumentList 'cd Backend/LightSensor_ConsumerSender; go run middleware/RabbitToSocketMiddleware.go'
Start-Process powershell -ArgumentList 'cd Backend/Notification_ConsumerSender; go run middleware/RabbitToSocketMiddleware.go'

Write-Host 'Todos los productores y consumidores han sido iniciados en nuevas ventanas de PowerShell.'
