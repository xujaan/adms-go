-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Host: mariadb
-- Generation Time: Jul 27, 2026 at 05:52 AM
-- Server version: 11.8.2-MariaDB-ubu2404
-- PHP Version: 8.2.29

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `java_adms`
--

-- --------------------------------------------------------

--
-- Table structure for table `attendances`
--

CREATE TABLE `attendances` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `sn` varchar(255) NOT NULL,
  `table` varchar(255) NOT NULL,
  `stamp` varchar(255) NOT NULL,
  `employee_id` int(11) NOT NULL,
  `timestamp` datetime NOT NULL,
  `status1` tinyint(1) DEFAULT NULL,
  `status2` tinyint(1) DEFAULT NULL,
  `status3` tinyint(1) DEFAULT NULL,
  `status4` tinyint(1) DEFAULT NULL,
  `status5` tinyint(1) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Triggers `attendances`
--
DELIMITER $$
CREATE TRIGGER `after_attendances_insert` AFTER INSERT ON `attendances` FOR EACH ROW BEGIN
    DECLARE emp_id UUID;
    DECLARE loc VARCHAR(100);

    -- Get the employee ID from employees table
    SELECT id_employees INTO emp_id
    FROM java_prod.employees
    WHERE employees_fingerprintid = NEW.employee_id
    LIMIT 1;

    -- Get the location from attendances_devices table
    SELECT location INTO loc
    FROM java_prod.attendances_devices
    WHERE device = NEW.sn
    LIMIT 1;

    -- Determine the checklog status
    IF NEW.status1 = 0 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'IN';
    ELSEIF NEW.status1 = 1 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'OUT';
    ELSEIF NEW.status1 = 4 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'LEMBUR IN';
    ELSEIF NEW.status1 = 5 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'LEMBUR OUT';
    ELSE
        SET @checklog = NULL;
    END IF;

    -- Insert into java_prod.attendances
    INSERT INTO java_prod.attendances (id_attendances, id_employees, id_absensi, attendances_checklog, timestamp, device, location)
    VALUES (UUID(), emp_id, NEW.employee_id, @checklog, NEW.timestamp, NEW.sn, loc);
END
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `after_attendances_update` AFTER UPDATE ON `attendances` FOR EACH ROW BEGIN
    DECLARE emp_id UUID;
    DECLARE loc VARCHAR(100);

    -- Get the employee ID from employees table
    SELECT id_employees INTO emp_id
    FROM java_prod.employees
    WHERE employees_fingerprintid = NEW.employee_id
    LIMIT 1;

    -- Get the location from attendances_devices table
    SELECT location INTO loc
    FROM java_prod.attendances_devices
    WHERE device = NEW.sn
    LIMIT 1;

    -- Determine the checklog status
    IF NEW.status1 = 0 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'IN';
    ELSEIF NEW.status1 = 1 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'OUT';
    ELSEIF NEW.status1 = 4 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'LEMBUR IN';
    ELSEIF NEW.status1 = 5 AND (NEW.status2 = 1 OR NEW.status2 = 0) THEN
        SET @checklog = 'LEMBUR OUT';
    ELSE
        SET @checklog = NULL;
    END IF;

    -- Insert into java_prod.attendances
    INSERT INTO java_prod.attendances (id_attendances, id_employees, id_absensi, attendances_checklog, timestamp, device, location, createdAt, updatedAt)
    VALUES (UUID(), emp_id, NEW.employee_id, @checklog, NEW.timestamp, NEW.sn, loc, NEW.timestamp, NEW.timestamp);
END
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `devices`
--

CREATE TABLE `devices` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `nama` varchar(255) DEFAULT NULL,
  `no_sn` varchar(255) NOT NULL,
  `lokasi` varchar(255) DEFAULT NULL,
  `online` datetime DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Triggers `devices`
--
DELIMITER $$
CREATE TRIGGER `after_devices_insert` AFTER INSERT ON `devices` FOR EACH ROW BEGIN
    INSERT INTO java_prod.attendances_devices (device, location, createdAt, updatedAt)
    VALUES (NEW.no_sn, NEW.lokasi, NOW(), NOW());
END
$$
DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `device_handshake_configs`
--

CREATE TABLE `device_handshake_configs` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `device_type` varchar(255) NOT NULL DEFAULT 'default',
  `stamp` int(11) NOT NULL DEFAULT 9999,
  `error_delay` int(11) NOT NULL DEFAULT 60,
  `delay` int(11) NOT NULL DEFAULT 30,
  `res_log_day` int(11) NOT NULL DEFAULT 18250,
  `res_log_del_count` int(11) NOT NULL DEFAULT 10000,
  `res_log_count` int(11) NOT NULL DEFAULT 50000,
  `trans_times` varchar(255) NOT NULL DEFAULT '00:00;14:05',
  `trans_interval` int(11) NOT NULL DEFAULT 1,
  `trans_flag` varchar(10) NOT NULL DEFAULT '1111000000',
  `time_zone` int(11) NOT NULL DEFAULT 7,
  `realtime` tinyint(1) NOT NULL DEFAULT 1,
  `encrypt` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `device_log`
--

CREATE TABLE `device_log` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `data` text NOT NULL,
  `tgl` date DEFAULT NULL,
  `sn` varchar(255) NOT NULL,
  `option` varchar(255) DEFAULT NULL,
  `url` varchar(255) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `error_log`
--

CREATE TABLE `error_log` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `data` text NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `failed_jobs`
--

CREATE TABLE `failed_jobs` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `uuid` varchar(255) NOT NULL,
  `connection` text NOT NULL,
  `queue` text NOT NULL,
  `payload` longtext NOT NULL,
  `exception` longtext NOT NULL,
  `failed_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `finger_log`
--

CREATE TABLE `finger_log` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `data` text NOT NULL,
  `url` text NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `migrations`
--

CREATE TABLE `migrations` (
  `id` int(10) UNSIGNED NOT NULL,
  `migration` varchar(255) NOT NULL,
  `batch` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `password_reset_tokens`
--

CREATE TABLE `password_reset_tokens` (
  `email` varchar(255) NOT NULL,
  `token` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `personal_access_tokens`
--

CREATE TABLE `personal_access_tokens` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `tokenable_type` varchar(255) NOT NULL,
  `tokenable_id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `token` varchar(64) NOT NULL,
  `abilities` text DEFAULT NULL,
  `last_used_at` timestamp NULL DEFAULT NULL,
  `expires_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `email_verified_at` timestamp NULL DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `remember_token` varchar(100) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Migration: Create webhooks table
-- Replaces hardcoded DB triggers with configurable webhook subscriptions

CREATE TABLE IF NOT EXISTS `webhooks` (
    `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
    `device_sn` varchar(255) NOT NULL DEFAULT '',
    `name` varchar(255) NOT NULL DEFAULT '',
    `url` varchar(1024) NOT NULL,
    `event` varchar(50) NOT NULL,
    `is_active` tinyint(1) NOT NULL DEFAULT 1,
    `created_at` timestamp NULL DEFAULT current_timestamp(),
    `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
    PRIMARY KEY (`id`),
    INDEX `idx_device_sn` (`device_sn`),
    INDEX `idx_event` (`event`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `attendances`
--
ALTER TABLE `attendances`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `devices`
--
ALTER TABLE `devices`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `devices_no_sn_unique` (`no_sn`);

--
-- Indexes for table `device_handshake_configs`
--
ALTER TABLE `device_handshake_configs`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `device_log`
--
ALTER TABLE `device_log`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `error_log`
--
ALTER TABLE `error_log`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `failed_jobs`
--
ALTER TABLE `failed_jobs`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `failed_jobs_uuid_unique` (`uuid`);

--
-- Indexes for table `finger_log`
--
ALTER TABLE `finger_log`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `migrations`
--
ALTER TABLE `migrations`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `password_reset_tokens`
--
ALTER TABLE `password_reset_tokens`
  ADD PRIMARY KEY (`email`);

--
-- Indexes for table `personal_access_tokens`
--
ALTER TABLE `personal_access_tokens`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `personal_access_tokens_token_unique` (`token`),
  ADD KEY `personal_access_tokens_tokenable_type_tokenable_id_index` (`tokenable_type`,`tokenable_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `users_email_unique` (`email`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `attendances`
--
ALTER TABLE `attendances`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `devices`
--
ALTER TABLE `devices`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `device_handshake_configs`
--
ALTER TABLE `device_handshake_configs`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `device_log`
--
ALTER TABLE `device_log`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `error_log`
--
ALTER TABLE `error_log`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `failed_jobs`
--
ALTER TABLE `failed_jobs`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `finger_log`
--
ALTER TABLE `finger_log`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `migrations`
--
ALTER TABLE `migrations`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `personal_access_tokens`
--
ALTER TABLE `personal_access_tokens`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
