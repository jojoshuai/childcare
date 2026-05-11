-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id              CHAR(36) PRIMARY KEY,
    username        VARCHAR(100) UNIQUE,
    password_hash   VARCHAR(255),
    wx_openid       VARCHAR(100) UNIQUE,
    nickname        VARCHAR(100) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 孩子表
CREATE TABLE IF NOT EXISTS children (
    id          CHAR(36) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    gender      ENUM('male', 'female') NOT NULL,
    birth_date  DATE NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 测量记录（体重/身高）
CREATE TABLE IF NOT EXISTS measurements (
    id          CHAR(36) PRIMARY KEY,
    child_id    CHAR(36) NOT NULL,
    type        ENUM('weight', 'height') NOT NULL,
    value       DECIMAL(10, 2) NOT NULL,
    measured_at TIMESTAMP NOT NULL,
    note        VARCHAR(500),
    created_by  CHAR(36) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE
);

-- 睡眠记录
CREATE TABLE IF NOT EXISTS sleep_records (
    id          CHAR(36) PRIMARY KEY,
    child_id    CHAR(36) NOT NULL,
    start_time  TIMESTAMP NOT NULL,
    end_time    TIMESTAMP NULL,
    woke_up     BOOLEAN NOT NULL DEFAULT FALSE,
    wake_count  INT NOT NULL DEFAULT 0,
    created_by  CHAR(36) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE
);

-- 饮食记录
CREATE TABLE IF NOT EXISTS diet_records (
    id             CHAR(36) PRIMARY KEY,
    child_id       CHAR(36) NOT NULL,
    food_name      VARCHAR(100) NOT NULL,
    food_type      ENUM('staple', 'vegetable', 'fruit', 'protein', 'dairy', 'snack') NOT NULL,
    amount_level   INT NOT NULL,
    record_time    TIMESTAMP NOT NULL,
    meal_group_id  CHAR(36) NULL,
    meal_type      ENUM('breakfast', 'lunch', 'dinner', 'snack') NULL,
    notes          VARCHAR(500),
    created_by     CHAR(36) NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE
);

-- 补剂记录
CREATE TABLE IF NOT EXISTS supplement_records (
    id               CHAR(36) PRIMARY KEY,
    child_id         CHAR(36) NOT NULL,
    supplement_name  VARCHAR(100) NOT NULL,
    dose             VARCHAR(100),
    taken_at         TIMESTAMP NOT NULL,
    created_by       CHAR(36) NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE
);

-- WHO 标准数据表
CREATE TABLE IF NOT EXISTS who_standards (
    id      INT AUTO_INCREMENT PRIMARY KEY,
    gender  ENUM('male', 'female') NOT NULL,
    type    ENUM('weight', 'height') NOT NULL,
    month   INT NOT NULL,
    p3      DECIMAL(10, 2) NOT NULL,
    p50     DECIMAL(10, 2) NOT NULL,
    p97     DECIMAL(10, 2) NOT NULL,
    UNIQUE KEY (gender, type, month)
);
