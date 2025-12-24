#!/usr/bin/env python3
"""
Інтерактивний скрипт для тестування класифікації звернень.
Дозволяє вводити текст і отримувати результати класифікації від ML сервісу.

Використання:
    python3 test-classification.py
    або
    ./test-classification.py

Налаштування:
    Можна змінити CLASSIFICATION_SERVICE_URL в коді або через змінну оточення.
"""

import requests
import json
import os
import sys
from typing import Optional

# Налаштування
CLASSIFICATION_SERVICE_URL = os.getenv(
    "CLASSIFICATION_SERVICE_URL", 
    "http://localhost:8000"
)

def classify_text(text: str) -> Optional[dict]:
    """
    Відправляє текст на класифікацію та повертає результат.
    
    Args:
        text: Текст для класифікації
        
    Returns:
        Словник з результатами класифікації або None у разі помилки
    """
    url = f"{CLASSIFICATION_SERVICE_URL}/classify"
    
    try:
        response = requests.post(
            url,
            json={"text": text},
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        if response.status_code == 200:
            return response.json()
        else:
            print(f"❌ Помилка: HTTP {response.status_code}")
            try:
                error_data = response.json()
                print(f"   Деталі: {error_data.get('detail', 'Невідома помилка')}")
            except:
                print(f"   Відповідь: {response.text}")
            return None
            
    except requests.exceptions.ConnectionError:
        print(f"❌ Помилка: Не вдалося підключитися до {url}")
        print("   Переконайтеся, що класифікаційний сервіс запущено.")
        return None
    except requests.exceptions.Timeout:
        print("❌ Помилка: Час очікування вичерпано")
        return None
    except Exception as e:
        print(f"❌ Несподівана помилка: {e}")
        return None


def format_result(result: dict) -> str:
    """
    Форматує результат класифікації для виведення.
    
    Args:
        result: Результат від API
        
    Returns:
        Відформатований рядок
    """
    output = []
    
    # Основна служба
    service = result.get("service", "Не визначено")
    confidence = result.get("confidence", 0.0)
    needs_moderation = result.get("needs_moderation", False)
    
    output.append("=" * 60)
    output.append("📋 РЕЗУЛЬТАТ КЛАСИФІКАЦІЇ")
    output.append("=" * 60)
    output.append(f"🎯 Служба: {service}")
    output.append(f"📊 Впевненість: {confidence:.2%}")
    
    if needs_moderation:
        output.append("⚠️  Потребує модерації: Так")
    else:
        output.append("✅ Потребує модерації: Ні")
    
    # Альтернативні варіанти
    top_alternatives = result.get("top_alternatives", [])
    if top_alternatives:
        output.append("")
        output.append("🔍 Альтернативні варіанти:")
        for i, alt in enumerate(top_alternatives[:5], 1):
            alt_service = alt.get("service", "Невідомо")
            alt_confidence = alt.get("confidence", 0.0)
            output.append(f"   {i}. {alt_service} ({alt_confidence:.2%})")
    
    output.append("=" * 60)
    
    return "\n".join(output)


def main():
    """Головна функція інтерактивного режиму."""
    print("=" * 60)
    print("🤖 ТЕСТУВАННЯ КЛАСИФІКАЦІЇ ЗВЕРНЕНЬ")
    print("=" * 60)
    print(f"🌐 Сервіс: {CLASSIFICATION_SERVICE_URL}")
    print("")
    print("Введіть текст звернення для класифікації.")
    print("Для виходу введіть 'exit', 'quit' або 'q'")
    print("Для очищення екрану введіть 'clear' або 'cls'")
    print("=" * 60)
    print("")
    
    while True:
        try:
            # Зчитуємо текст від користувача
            text = input("\n📝 Введіть текст звернення: ").strip()
            
            # Перевірка на вихід
            if text.lower() in ['exit', 'quit', 'q']:
                print("\n👋 До побачення!")
                break
            
            # Перевірка на очищення
            if text.lower() in ['clear', 'cls']:
                os.system('clear' if os.name != 'nt' else 'cls')
                continue
            
            # Перевірка на порожній ввід
            if not text:
                print("⚠️  Будь ласка, введіть текст для класифікації.")
                continue
            
            # Відправляємо на класифікацію
            print("\n⏳ Обробка...")
            result = classify_text(text)
            
            if result:
                print("\n" + format_result(result))
            else:
                print("\n❌ Не вдалося отримати результат класифікації.")
            
        except KeyboardInterrupt:
            print("\n\n👋 Перервано користувачем. До побачення!")
            break
        except EOFError:
            print("\n\n👋 До побачення!")
            break
        except Exception as e:
            print(f"\n❌ Несподівана помилка: {e}")


if __name__ == "__main__":
    main()

