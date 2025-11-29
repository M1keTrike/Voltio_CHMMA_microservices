-- ==============================================================================
-- VOLTIO BACKEND - SQL para Crear Tabla automation_rules
-- Ejecutar este SQL en tu PostgreSQL existente (postgres-local)
-- ==============================================================================

-- Conectar a tu base de datos:
-- docker exec -it postgres-local psql -U mike -d voltio_db

-- ==============================================================================
-- Crear tabla de reglas de automatización
-- ==============================================================================
CREATE TABLE IF NOT EXISTS automation_rules (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    trigger_device_mac VARCHAR(17) NOT NULL,
    action_device_mac VARCHAR(17) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    trigger_metric VARCHAR(50) NOT NULL,
    comparison_operator VARCHAR(20),
    threshold_value DECIMAL(10, 2),
    action_capability_id INTEGER NOT NULL,
    action_payload VARCHAR(10) NOT NULL,
    active_time_start TIME,
    active_time_end TIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==============================================================================
-- Crear índices para mejorar el rendimiento
-- ==============================================================================
CREATE INDEX IF NOT EXISTS idx_automation_rules_trigger_mac 
    ON automation_rules(trigger_device_mac);

CREATE INDEX IF NOT EXISTS idx_automation_rules_is_active 
    ON automation_rules(is_active);

CREATE INDEX IF NOT EXISTS idx_automation_rules_user_id 
    ON automation_rules(user_id);

-- ==============================================================================
-- Insertar reglas de ejemplo (opcional)
-- ==============================================================================
INSERT INTO automation_rules (
    user_id, 
    trigger_device_mac, 
    action_device_mac, 
    name, 
    is_active, 
    trigger_metric, 
    comparison_operator, 
    threshold_value, 
    action_capability_id, 
    action_payload,
    active_time_start,
    active_time_end
) VALUES 
-- Regla 1: Encender luz con movimiento
(
    1, 
    'd8:3a:dd:09:ff:99', 
    'CC:DB:A7:2F:AE:B0', 
    'Encender luz con movimiento', 
    true, 
    'motion', 
    'EQUAL', 
    1.0, 
    1, 
    'ON',
    '08:00:00',
    '20:00:00'
),
-- Regla 2: Apagar luz sin movimiento por 30 min
(
    1, 
    'd8:3a:dd:09:ff:99', 
    'CC:DB:A7:2F:AE:B0', 
    'Apagar luz sin movimiento por 30 min', 
    true, 
    'motion_timeout', 
    'GREATER_THAN', 
    1800.0, 
    1, 
    'OFF',
    NULL,
    NULL
),
-- Regla 3: Encender si temperatura > 25°C
(
    1, 
    'aa:bb:cc:dd:ee:ff', 
    'CC:DB:A7:2F:AE:B0', 
    'Encender si temperatura > 25°C', 
    true, 
    'temperature', 
    'GREATER_THAN', 
    25.0, 
    1, 
    'ON',
    NULL,
    NULL
)
ON CONFLICT DO NOTHING;

-- ==============================================================================
-- Verificar que la tabla se creó correctamente
-- ==============================================================================
-- Ver estructura de la tabla
\d automation_rules

-- Ver reglas insertadas
SELECT id, name, trigger_metric, is_active FROM automation_rules;

-- ==============================================================================
-- Comentarios de documentación
-- ==============================================================================
COMMENT ON TABLE automation_rules IS 'Reglas de automatización IoT configurables por usuario';
COMMENT ON COLUMN automation_rules.trigger_metric IS 'Métrica del sensor: motion, temperature, humidity, lux, voltage, current, power, energy, frequency, pf, workday_start, workday_end, motion_timeout';
COMMENT ON COLUMN automation_rules.comparison_operator IS 'Operador: GREATER_THAN, LESS_THAN, EQUAL, NOT_EQUAL';
COMMENT ON COLUMN automation_rules.action_capability_id IS '1=RELAY_CONTROL, 2=INFRARED_EMITTER';

-- ==============================================================================
-- COMANDOS ÚTILES
-- ==============================================================================

-- Ver todas las reglas activas:
-- SELECT * FROM automation_rules WHERE is_active = true;

-- Desactivar una regla:
-- UPDATE automation_rules SET is_active = false WHERE id = 1;

-- Eliminar todas las reglas de ejemplo:
-- DELETE FROM automation_rules;

-- Salir de psql:
-- \q
