#include <iostream>
#include <vector>
#include <map>
#include <cmath>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <thread>
#include <future>
#include <mutex>
#include <chrono>
#include <cstring>

#define PORT 9001
#define BROADCAST_IP "255.255.255.255"
#define TIMEOUT_SEC 3

std::mutex mutex;
std::map<std::pair<int, std::string>, bool> server_status;

void broadcast_search(int sock, struct sockaddr_in &broadcast_addr);
void receive_server_responses(int sock);
double integrate(double a, double b, int n, const std::vector<std::pair<int, std::string>> &servers);

int main() {
    std::this_thread::sleep_for(std::chrono::seconds(5));  // Задержка
    std::cout << "START\n";
    std::cerr << "START\n";

    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    struct sockaddr_in broadcast_addr;
    broadcast_addr.sin_family = AF_INET;
    broadcast_addr.sin_port = htons(PORT);
    broadcast_addr.sin_addr.s_addr = INADDR_BROADCAST;

    int broadcast_enable = 1;
    setsockopt(sock, SOL_SOCKET, SO_BROADCAST, &broadcast_enable, sizeof(broadcast_enable));

    // Шаг 1: Поиск доступных серверов
    broadcast_search(sock, broadcast_addr);
    
    // Ожидание ответов от серверов
    receive_server_responses(sock);
    
    double a = 0.0, b = 10.0;
    int n = 1000;
    std::vector<std::pair<int, std::string>> servers;
    for (const auto &entry : server_status) {
        if (entry.second) servers.push_back(entry.first);
    }
    
    if (servers.empty()) {
        std::cerr << "Нет доступных серверов." << std::endl;
        close(sock);
        return -1;
    }

    double result = integrate(a, b, n, servers);
    std::cout << "Результат интегрирования: " << result << std::endl;
    
    close(sock);
    return 0;
}

void broadcast_search(int sock, struct sockaddr_in &broadcast_addr) {
    const char *message = "MASTER_DISCOVERY";
    sendto(sock, message, strlen(message), 0, (struct sockaddr *)&broadcast_addr, sizeof(broadcast_addr));
}

void receive_server_responses(int sock) {
    fd_set readfds;
    struct timeval timeout;
    timeout.tv_sec = TIMEOUT_SEC;
    timeout.tv_usec = 0;
    
    FD_ZERO(&readfds);
    FD_SET(sock, &readfds);
    
    if (select(sock + 1, &readfds, NULL, NULL, &timeout) > 0) {
        struct sockaddr_in server_addr;
        socklen_t addr_len = sizeof(server_addr);
        char buffer[256];
        
        recvfrom(sock, buffer, sizeof(buffer), 0, (struct sockaddr *)&server_addr, &addr_len);
        int server_id = ntohs(server_addr.sin_port); // допустим, серверы используют уникальный порт как идентификатор
        std::cout << "Сервер найден: " << inet_ntoa(server_addr.sin_addr) << std::endl;
        
        std::string host = inet_ntoa(server_addr.sin_addr);
        std::lock_guard<std::mutex> lock(mutex);
        server_status[{server_id, host}] = true;
    }
}

double integrate(double a, double b, int n, const std::vector<std::pair<int, std::string>> &servers) {
    double total_result = 0.0;
    int segment_size = n / servers.size();
    double segment_length = (b - a) / servers.size();

    std::vector<std::thread> jobs;
    std::vector<std::future<double>> futures;
    std::vector<std::promise<double>> promies(servers.size());

    for (size_t i = 0; i < servers.size(); ++i) {
        futures.emplace_back(std::move(promies[i].get_future()));
    }

    for (size_t i = 0; i < servers.size(); ++i) {
        jobs.emplace_back([&, server_id = i] {
            int sock = socket(AF_INET, SOCK_DGRAM, 0);
            struct sockaddr_in server_addr;

            double segment_start = a + server_id * segment_length;
            double segment_end = segment_start + segment_length;

            server_addr.sin_family = AF_INET;
            server_addr.sin_port = htons(servers[server_id].first);
            inet_pton(AF_INET, servers[server_id].second.c_str(), &server_addr.sin_addr); // Используем локальный IP для примера

            char message[256];
            snprintf(message, sizeof(message), "INTEGRATE %f %f %d", segment_start, segment_end, segment_size);
            sendto(sock, message, strlen(message), 0, (struct sockaddr *)&server_addr, sizeof(server_addr));

            struct sockaddr_in from_addr;
            socklen_t from_len = sizeof(from_addr);
            char buffer[256];
            recvfrom(sock, buffer, sizeof(buffer), 0, (struct sockaddr *)&from_addr, &from_len);
        
            double partial_result = atof(buffer);
            promies[server_id].set_value(partial_result);
            close(sock);
        });
    }

    for (size_t i = 0; i < servers.size(); ++i) {
        jobs[i].join();
        total_result += futures[i].get();
    }

    return total_result;
}
