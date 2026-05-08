-- 睡眠记录
CREATE TABLE IF NOT EXISTS sleep_records (
    id         CHAR(36)    NOT NULL,
    child_id   CHAR(36)    NOT NULL,
    start_time DATETIME    NOT NULL,
    end_time   DATETIME    NULL,
    woke_up    TINYINT(1)  NOT NULL DEFAULT 0,
    wake_count INT         NOT NULL DEFAULT 0,
    created_by CHAR(36)    NOT NULL,
    created_at DATETIME    NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_sleep_records_child FOREIGN KEY (child_id) REFERENCES children (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 饮食记录
CREATE TABLE IF NOT EXISTS diet_records (
    id           CHAR(36)     NOT NULL,
    child_id     CHAR(36)     NOT NULL,
    food_name    VARCHAR(100) NOT NULL,
    food_type    ENUM('staple','vegetable','fruit','protein','dairy','snack') NOT NULL,
    amount_level TINYINT      NOT NULL COMMENT '1=少,2=正常,3=多',
    record_time  DATETIME     NOT NULL,
    notes        VARCHAR(500) NULL,
    created_by   CHAR(36)     NOT NULL,
    created_at   DATETIME     NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_diet_records_child FOREIGN KEY (child_id) REFERENCES children (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 补剂打卡
CREATE TABLE IF NOT EXISTS supplement_records (
    id              CHAR(36)    NOT NULL,
    child_id        CHAR(36)    NOT NULL,
    supplement_name VARCHAR(50) NOT NULL,
    dose            VARCHAR(50) NULL,
    taken_at        DATETIME    NOT NULL,
    created_by      CHAR(36)    NOT NULL,
    created_at      DATETIME    NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_supplement_records_child FOREIGN KEY (child_id) REFERENCES children (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
