# Базовый образ для компиляции
FROM gcc:latest AS builder

# Устанавливаем необходимые зависимости
RUN apt-get update && apt-get install -y build-essential && apt-get install -y libstdc++6

# Копируем файлы исходного кода
COPY master.cpp /app/master.cpp
COPY server.cpp /app/server.cpp

# Компилируем мастер и сервер
WORKDIR /app
RUN g++ -static-libstdc++ master.cpp -o master -lpthread
RUN g++ -static-libstdc++ server.cpp -o server -lpthread

# Базовый образ для запуска
FROM debian:latest

# Копируем скомпилированные файлы
COPY --from=builder /app/master /app/master
COPY --from=builder /app/server /app/server

# Открываем порты
EXPOSE 9000

# Команда для запуска контейнера
CMD ["/bin/bash"]
