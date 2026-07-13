import sqlite3
import random
from datetime import datetime, timedelta

db_path = r"C:\app\data\app.db"
conn = sqlite3.connect(db_path)
cur = conn.cursor()

# 获取所有宝宝
cur.execute("SELECT id, name, birth_date, user_id FROM babies")
babies = cur.fetchall()
print(f"找到 {len(babies)} 个宝宝:")
for b in babies:
    print(f"  ID={b[0]}, 姓名={b[1]}, 出生={b[2]}, 用户={b[3]}")

if not babies:
    print("没有宝宝，无法添加测试数据")
    conn.close()
    exit()

feeding_types = ['breast', 'bottle', 'mixed']
diaper_types = ['wet', 'dirty', 'both']
now = datetime.now()

for baby_id, name, birth_date, user_id in babies:
    print(f"\n为 {name} (ID={baby_id}) 添加测试数据...")

    for day_offset in range(7, -1, -1):
        day = now - timedelta(days=day_offset)

        # 每天 5-8 次喂奶
        feed_count = random.randint(5, 8)
        for i in range(feed_count):
            hour = random.randint(6, 22)
            minute = random.randint(0, 59)
            second = random.randint(0, 59)
            occurred = day.replace(hour=hour, minute=minute, second=second)
            ftype = random.choice(feeding_types)
            if ftype == 'breast':
                amount = random.randint(80, 150)
                duration = random.randint(10, 25)
            else:
                amount = random.randint(120, 200)
                duration = 0
            cur.execute(
                "INSERT INTO feeding_records (baby_id, user_id, type, amount_ml, duration_minutes, note, occurred_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
                (baby_id, user_id, ftype, amount, duration, "", occurred.strftime("%Y-%m-%d %H:%M:%S"), occurred.strftime("%Y-%m-%d %H:%M:%S"))
            )

        # 每天 4-6 次尿布
        diaper_count = random.randint(4, 6)
        for i in range(diaper_count):
            hour = random.randint(6, 22)
            minute = random.randint(0, 59)
            second = random.randint(0, 59)
            occurred = day.replace(hour=hour, minute=minute, second=second)
            dtype = random.choice(diaper_types)
            cur.execute(
                "INSERT INTO diaper_records (baby_id, user_id, type, note, occurred_at, created_at) VALUES (?, ?, ?, ?, ?, ?)",
                (baby_id, user_id, dtype, "", occurred.strftime("%Y-%m-%d %H:%M:%S"), occurred.strftime("%Y-%m-%d %H:%M:%S"))
            )

    # 今日额外添加几条记录确保有数据
    today_feeds = random.randint(2, 4)
    for i in range(today_feeds):
        hour = random.randint(max(6, now.hour - 6), now.hour)
        minute = random.randint(0, 59)
        second = random.randint(0, 59)
        occurred = now.replace(hour=hour, minute=minute, second=second)
        cur.execute(
            "INSERT INTO feeding_records (baby_id, user_id, type, amount_ml, duration_minutes, note, occurred_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
            (baby_id, user_id, 'bottle', random.randint(120, 180), 0, "测试", occurred.strftime("%Y-%m-%d %H:%M:%S"), occurred.strftime("%Y-%m-%d %H:%M:%S"))
        )

conn.commit()

# 统计添加结果
for baby_id, name, birth_date, user_id in babies:
    cur.execute("SELECT COUNT(*) FROM feeding_records WHERE baby_id=?", (baby_id,))
    feeds = cur.fetchone()[0]
    cur.execute("SELECT COUNT(*) FROM diaper_records WHERE baby_id=?", (baby_id,))
    diapers = cur.fetchone()[0]
    cur.execute("SELECT SUM(amount_ml) FROM feeding_records WHERE baby_id=?", (baby_id,))
    total_ml = cur.fetchone()[0] or 0
    print(f"  {name}: 喂奶 {feeds} 条, 尿布 {diapers} 条, 总奶量 {total_ml} ml")

conn.close()
print("\n测试数据添加完成！")
