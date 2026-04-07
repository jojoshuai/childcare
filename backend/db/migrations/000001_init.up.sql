CREATE TABLE IF NOT EXISTS families (
    id         CHAR(36)     NOT NULL,
    name       VARCHAR(100) NOT NULL,
    created_at DATETIME     NOT NULL,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS users (
    id            CHAR(36)              NOT NULL,
    family_id     CHAR(36)              NULL,
    username      VARCHAR(50)           NULL,
    password_hash VARCHAR(255)          NULL,
    wx_openid     VARCHAR(100)          NULL,
    nickname      VARCHAR(50)           NOT NULL,
    role          ENUM('owner','member') NULL,
    created_at    DATETIME              NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_username  (username),
    UNIQUE KEY uq_users_wx_openid (wx_openid),
    CONSTRAINT fk_users_family FOREIGN KEY (family_id) REFERENCES families (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS children (
    id         CHAR(36)               NOT NULL,
    family_id  CHAR(36)               NOT NULL,
    name       VARCHAR(50)            NOT NULL,
    gender     ENUM('male','female')  NOT NULL,
    birth_date DATE                   NOT NULL,
    created_at DATETIME               NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_children_family FOREIGN KEY (family_id) REFERENCES families (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS measurements (
    id          CHAR(36)                                              NOT NULL,
    child_id    CHAR(36)                                             NOT NULL,
    type        ENUM('weight','height','head_circumference')         NOT NULL,
    value       DECIMAL(6,2)                                         NOT NULL,
    measured_at DATE                                                 NOT NULL,
    note        VARCHAR(500)                                         NULL,
    created_by  CHAR(36)                                             NOT NULL,
    created_at  DATETIME                                             NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_measurements_child FOREIGN KEY (child_id) REFERENCES children (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS invite_codes (
    id         CHAR(36)    NOT NULL,
    family_id  CHAR(36)    NOT NULL,
    code       CHAR(6)     NOT NULL,
    expires_at DATETIME    NOT NULL,
    used       TINYINT(1)  NOT NULL DEFAULT 0,
    created_by CHAR(36)    NOT NULL,
    created_at DATETIME    NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_invite_codes_code (code),
    CONSTRAINT fk_invite_codes_family FOREIGN KEY (family_id) REFERENCES families (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
