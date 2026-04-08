// miniprogram/utils/util.js

/**
 * 计算从 birthDate 到 endDate（默认今天）的月龄（整数，向下取整）
 * @param {string} birthDate - ISO 日期字符串，如 '2023-01-15'
 * @param {string} [endDate]  - ISO 日期字符串，不传则用今天
 * @returns {number}
 */
function calcAgeMonths(birthDate, endDate) {
  const end = endDate ? new Date(endDate) : new Date()
  const diffMs = end.getTime() - new Date(birthDate).getTime()
  const diffDays = diffMs / (1000 * 60 * 60 * 24)
  return Math.floor(diffDays / 30.4375)
}

/**
 * 将 birthDate 转为友好月龄字符串，如 "8个月" 或 "1岁3个月"
 * @param {string} birthDate
 * @returns {string}
 */
function ageLabel(birthDate) {
  const months = calcAgeMonths(birthDate)
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

/**
 * 今天的 YYYY-MM-DD 字符串
 * @returns {string}
 */
function today() {
  return new Date().toISOString().slice(0, 10)
}

module.exports = { calcAgeMonths, ageLabel, today }
