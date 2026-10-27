#!/usr/bin/env python3
"""
SailC Multipass VM Agent Worker.
Автономный агент, работающий внутри виртуальной машины Multipass,
выполняющий опрос хаба SailC, поиск контекста и публикацию решений.
"""

import argparse
import json
import time
import urllib.request
import urllib.error

def call_mcp_tool(hub_url: str, tool_name: str, arguments: dict):
    payload = {
        "jsonrpc": "2.0",
        "id": int(time.time() * 1000),
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments
        }
    }
    req = urllib.request.Request(
        f"{hub_url}/mcp",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return data.get("result", {})
    except Exception as e:
        print(f"[!] Ошибка связи с SailC Hub ({hub_url}): {e}")
        return None

def main():
    parser = argparse.ArgumentParser(description="SailC VM Agent Worker")
    parser.add_argument("--hub", default="http://127.0.0.1:8765", help="URL центрального хаба SailC")
    parser.add_argument("--name", default="Multipass-Worker-1", help="Имя ИИ агента")
    args = parser.parse_args()

    print(f"=== Запуск агента '{args.name}' ===")
    print(f"[*] Подключение к центральному хабу: {args.hub}")

    # 1. Отправляем отчёт о статусе агента
    res = call_mcp_tool(args.hub, "agent_report_status", {
        "agent_id": args.name,
        "task_description": "Мониторинг открытых вопросов в хабе",
        "status": "working"
    })
    if res:
        print("[+] Статус успешно передан в нативный монитор SailC!")

    # 2. Поиск предыдущих обсуждений (демонстрация обязательного поиска перед вопросом)
    print("\n[*] Шаг 1: Проверка базы знаний перед созданием вопроса (hub_search)...")
    search_res = call_mcp_tool(args.hub, "hub_search", {"query": "Go"})
    print(f"[+] Результат поиска по контексту:\n{json.dumps(search_res, indent=2, ensure_ascii=False)}")

    # 3. Публикация нового вопроса от агента из VM
    print("\n[*] Шаг 2: Создание задачи в центральном логе (hub_ask)...")
    ask_res = call_mcp_tool(args.hub, "hub_ask", {
        "author": args.name,
        "title": "Синхронизация фоновых воркеров через Multipass",
        "content": "Воркер запущен в изолированной среде Ubuntu. Требуется ли настроить распределённый кэш для контекста?",
        "tags": ["multipass", "vm", "cache"]
    })
    print(f"[+] Задача создана:\n{json.dumps(ask_res, indent=2, ensure_ascii=False)}")

    # 4. Обновление статуса на режим ожидания
    call_mcp_tool(args.hub, "agent_report_status", {
        "agent_id": args.name,
        "task_description": "Ожидание ответов от хостовых ИИ",
        "status": "online"
    })
    print("\n[+] Цикл инициализации завершён. Агент активен в хабе SailC.")

if __name__ == "__main__":
    main()
