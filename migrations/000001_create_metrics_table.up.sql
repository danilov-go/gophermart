-- Таблица пользователей и их текущего баланса
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    current DOUBLE PRECISION NOT NULL DEFAULT 0.00,
    withdrawn DOUBLE PRECISION NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица загруженных заказов для расчета баллов лояльности
CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'NEW',
    accrual DOUBLE PRECISION NOT NULL DEFAULT 0.00,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Связь заказов с пользователем
ALTER TABLE orders 
    ADD CONSTRAINT fk_user 
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Таблица истории списаний баллов пользователей
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,             
    user_id INTEGER NOT NULL,           
    order_number VARCHAR(255) NOT NULL,  
    sum DOUBLE PRECISION NOT NULL DEFAULT 0.00,         
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() 
);

-- Связь записей списаний с пользователем
ALTER TABLE withdrawals 
    ADD CONSTRAINT fk_withdraw_user 
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE;