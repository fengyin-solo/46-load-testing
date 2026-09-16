package service

import (
	"strconv"
	"strings"
	"time"
)

// cronParser 解析 5 段式 cron 表达式并计算下次运行时间。
// 格式：分 时 日 月 周，各字段支持 *、数字、逗号、范围 a-b、步进 /n。
type cronParser struct {
	minutes    map[int]bool
	hours      map[int]bool
	days       map[int]bool
	months     map[int]bool
	weekdays   map[int]bool
}

// parseCron 解析 cron 表达式。
func parseCron(expr string) (*cronParser, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, errInvalidCron
	}
	minutes, err := parseCronField(fields[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hours, err := parseCronField(fields[1], 0, 23)
	if err != nil {
		return nil, err
	}
	days, err := parseCronField(fields[2], 1, 31)
	if err != nil {
		return nil, err
	}
	months, err := parseCronField(fields[3], 1, 12)
	if err != nil {
		return nil, err
	}
	weekdays, err := parseWeekdayField(fields[4])
	if err != nil {
		return nil, err
	}
	return &cronParser{
		minutes:  minutes,
		hours:    hours,
		days:     days,
		months:   months,
		weekdays: weekdays,
	}, nil
}

var errInvalidCron = &cronError{"cron 表达式不合法"}

type cronError struct{ msg string }

func (e *cronError) Error() string { return e.msg }

// parseCronField 解析普通数值字段，返回允许的值集合。
func parseCronField(field string, min, max int) (map[int]bool, error) {
	result := make(map[int]bool)
	parts := strings.Split(field, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errInvalidCron
		}
		values, err := expandCronPart(part, min, max)
		if err != nil {
			return nil, err
		}
		for _, v := range values {
			result[v] = true
		}
	}
	return result, nil
}

// parseWeekdayField 解析周字段，7 映射为周日 0。
func parseWeekdayField(field string) (map[int]bool, error) {
	result := make(map[int]bool)
	parts := strings.Split(field, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errInvalidCron
		}
		values, err := expandCronPart(part, 0, 7)
		if err != nil {
			return nil, err
		}
		for _, v := range values {
			if v == 7 {
				v = 0
			}
			result[v] = true
		}
	}
	return result, nil
}

// expandCronPart 展开单个 cron 片段（含 *、范围、步进）。
func expandCronPart(part string, min, max int) ([]int, error) {
	// 步进
	step := 1
	base := part
	if i := strings.Index(part, "/"); i >= 0 {
		base = part[:i]
		s, err := strconv.Atoi(part[i+1:])
		if err != nil || s <= 0 {
			return nil, errInvalidCron
		}
		step = s
	}
	start, end := min, max
	switch {
	case base == "*":
		// 全范围
	case strings.Contains(base, "-"):
		seg := strings.SplitN(base, "-", 2)
		a, err := strconv.Atoi(seg[0])
		if err != nil {
			return nil, errInvalidCron
		}
		b, err := strconv.Atoi(seg[1])
		if err != nil {
			return nil, errInvalidCron
		}
		if a < min || b > max || a > b {
			return nil, errInvalidCron
		}
		start, end = a, b
	default:
		n, err := strconv.Atoi(base)
		if err != nil || n < min || n > max {
			return nil, errInvalidCron
		}
		start, end = n, n
	}
	values := make([]int, 0, (end-start)/step+1)
	for v := start; v <= end; v += step {
		values = append(values, v)
	}
	return values, nil
}

// match 判断时间是否命中 cron 表达式。
func (c *cronParser) match(t time.Time) bool {
	return c.minutes[t.Minute()] &&
		c.hours[t.Hour()] &&
		c.months[int(t.Month())] &&
		c.days[t.Day()] &&
		c.weekdays[int(t.Weekday())]
}

// next 计算 from 之后（不含）的下一次匹配时间。
func (c *cronParser) next(from time.Time) time.Time {
	// 从下一分钟开始逐分钟探测，上限 2 年。
	candidate := from.Truncate(time.Minute).Add(time.Minute)
	limit := from.AddDate(2, 0, 0)
	for candidate.Before(limit) {
		if c.match(candidate) {
			return candidate
		}
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}
}

// NextRunAfter 计算 cron 表达式在指定时间之后的下次运行时间。
func NextRunAfter(expr string, after time.Time) (time.Time, error) {
	p, err := parseCron(expr)
	if err != nil {
		return time.Time{}, err
	}
	next := p.next(after)
	if next.IsZero() {
		return time.Time{}, errInvalidCron
	}
	return next, nil
}
