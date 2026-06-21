-- 粮库熏蒸和温控管理系统数据库 Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. 用户表（角色：保管员keeper、安全员safety_officer、值班员duty_officer）
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('keeper', 'safety_officer', 'duty_officer', 'admin')),
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 2. 仓房表
CREATE TABLE IF NOT EXISTS granaries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(200),
    capacity NUMERIC(12,2),
    grain_type VARCHAR(50),
    grain_variety VARCHAR(100),
    grain_weight NUMERIC(12,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (status IN ('normal', 'fumigating', 'ventilating', 'sealed', 'abnormal')),
    keeper_id UUID REFERENCES users(id),
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. 传感器表
CREATE TABLE IF NOT EXISTS sensors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('temperature', 'humidity', 'gas_ph3', 'gas_h2s', 'co2', 'o2')),
    location_desc VARCHAR(200),
    position_x NUMERIC(8,2),
    position_y NUMERIC(8,2),
    position_z NUMERIC(8,2),
    unit VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 4. 粮情记录（保管员录入）
CREATE TABLE IF NOT EXISTS grain_condition_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    recorder_id UUID NOT NULL REFERENCES users(id),
    record_time TIMESTAMPTZ NOT NULL,
    avg_temperature NUMERIC(6,2),
    max_temperature NUMERIC(6,2),
    min_temperature NUMERIC(6,2),
    avg_humidity NUMERIC(6,2),
    grain_level NUMERIC(8,2),
    pest_found BOOLEAN DEFAULT false,
    mold_found BOOLEAN DEFAULT false,
    abnormal_areas JSONB,
    weather_condition VARCHAR(100),
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 5. 传感器读数表
CREATE TABLE IF NOT EXISTS sensor_readings (
    id BIGSERIAL PRIMARY KEY,
    sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    reading_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    value NUMERIC(12,4) NOT NULL,
    is_abnormal BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_sensor_readings_granary_time ON sensor_readings(granary_id, reading_time DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_readings_sensor_time ON sensor_readings(sensor_id, reading_time DESC);

-- 6. 熏蒸方案表
CREATE TABLE IF NOT EXISTS fumigation_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    plan_no VARCHAR(50) UNIQUE NOT NULL,
    plan_title VARCHAR(200) NOT NULL,
    creator_id UUID NOT NULL REFERENCES users(id),
    chemical_type VARCHAR(50),
    chemical_name VARCHAR(100),
    dosage NUMERIC(10,2),
    dosage_unit VARCHAR(20),
    target_concentration NUMERIC(10,4),
    plan_start_time TIMESTAMPTZ,
    plan_end_time TIMESTAMPTZ,
    expected_seal_hours INTEGER,
    reason TEXT,
    people_cleared BOOLEAN DEFAULT false,
    people_cleared_time TIMESTAMPTZ,
    people_cleared_by UUID REFERENCES users(id),
    status VARCHAR(30) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending_approval', 'approved', 'rejected', 'in_progress', 'completed', 'cancelled')),
    approver_id UUID REFERENCES users(id),
    approval_remark TEXT,
    approved_at TIMESTAMPTZ,
    detection_interval_hours INTEGER DEFAULT 4,
    next_detection_time TIMESTAMPTZ,
    safety_confirmed BOOLEAN DEFAULT false,
    safety_confirmed_at TIMESTAMPTZ,
    safety_confirmed_by UUID REFERENCES users(id),
    safety_confirm_remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 7. 熏蒸执行记录表
CREATE TABLE IF NOT EXISTS fumigation_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id UUID NOT NULL REFERENCES fumigation_plans(id) ON DELETE CASCADE,
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    operator_id UUID NOT NULL REFERENCES users(id),
    actual_start_time TIMESTAMPTZ,
    actual_end_time TIMESTAMPTZ,
    chemical_actual_dosage NUMERIC(10,2),
    concentration_readings JSONB,
    weather_during VARCHAR(100),
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 8. 通风解封记录表
CREATE TABLE IF NOT EXISTS unseal_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    fumigation_plan_id UUID REFERENCES fumigation_plans(id),
    recorder_id UUID NOT NULL REFERENCES users(id),
    unseal_type VARCHAR(20) NOT NULL CHECK (unseal_type IN ('ventilation', 'unseal')),
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    weather_condition VARCHAR(100),
    is_safe BOOLEAN DEFAULT false,
    final_gas_readings JSONB,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 9. 气体检测记录表
CREATE TABLE IF NOT EXISTS gas_detection_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    unseal_id UUID REFERENCES unseal_records(id) ON DELETE CASCADE,
    detector_id UUID NOT NULL REFERENCES users(id),
    detection_time TIMESTAMPTZ NOT NULL,
    gas_type VARCHAR(30) NOT NULL,
    concentration NUMERIC(10,4) NOT NULL,
    safe_limit NUMERIC(10,4) NOT NULL,
    is_safe BOOLEAN DEFAULT false,
    detection_points JSONB,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 10. 翻仓建议表
CREATE TABLE IF NOT EXISTS grain_turnover_suggestions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    granary_id UUID NOT NULL REFERENCES granaries(id) ON DELETE CASCADE,
    source_record_id UUID REFERENCES grain_condition_records(id),
    suggestion_no VARCHAR(50) UNIQUE,
    abnormal_area_desc TEXT,
    temperature_anomaly JSONB,
    suggestion_content TEXT NOT NULL,
    priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'ignored')),
    handler_id UUID REFERENCES users(id),
    handled_at TIMESTAMPTZ,
    handle_remark TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 11. 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    module VARCHAR(50),
    target_id UUID,
    detail TEXT,
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 插入初始用户数据
INSERT INTO users (id, username, full_name, role, password_hash) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin', '系统管理员', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('00000000-0000-0000-0000-000000000002', 'keeper01', '张保管员', 'keeper', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('00000000-0000-0000-0000-000000000003', 'safety01', '李安全员', 'safety_officer', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'),
    ('00000000-0000-0000-0000-000000000004', 'duty01', '王值班员', 'duty_officer', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy')
ON CONFLICT (id) DO NOTHING;

-- 插入示例仓房
INSERT INTO granaries (id, code, name, location, capacity, grain_type, grain_weight, keeper_id, status) VALUES
    ('10000000-0000-0000-0000-000000000001', 'A-01', '一号仓', '东区A栋1层', 5000.00, '小麦', 4800.00, '00000000-0000-0000-0000-000000000002', 'normal'),
    ('10000000-0000-0000-0000-000000000002', 'A-02', '二号仓', '东区A栋2层', 6000.00, '玉米', 5500.00, '00000000-0000-0000-0000-000000000002', 'normal'),
    ('10000000-0000-0000-0000-000000000003', 'B-01', '三号仓', '西区B栋1层', 4500.00, '稻谷', 4200.00, '00000000-0000-0000-0000-000000000002', 'normal')
ON CONFLICT (id) DO NOTHING;

-- 插入示例传感器
INSERT INTO sensors (id, granary_id, code, type, location_desc, position_x, position_y, position_z, unit) VALUES
    ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'A01-T01', 'temperature', '上层-东北', 2.00, 2.00, 4.50, '°C'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 'A01-T02', 'temperature', '中层-中心', 5.00, 5.00, 3.00, '°C'),
    ('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 'A01-T03', 'temperature', '下层-西南', 8.00, 8.00, 1.00, '°C'),
    ('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', 'A01-H01', 'humidity', '上层', 5.00, 5.00, 4.50, '%RH'),
    ('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', 'A01-G01', 'gas_ph3', '仓内中心', 5.00, 5.00, 3.00, 'ppm')
ON CONFLICT (id) DO NOTHING;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_granaries_updated_at BEFORE UPDATE ON granaries FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_fumigation_plans_updated_at BEFORE UPDATE ON fumigation_plans FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_grain_turnover_suggestions_updated_at BEFORE UPDATE ON grain_turnover_suggestions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
