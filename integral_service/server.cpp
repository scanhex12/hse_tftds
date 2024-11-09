#include <iostream>
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <cstring>
#include <cmath>

#define PORT 9001

double compute_integral(double (*f)(double), double a, double b, int n);
double function(double x);

int main() {
    std::cout << "SERVER" << std::endl;
    std::cerr << "SERVER" << std::endl;
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    struct sockaddr_in server_addr;
    
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(PORT);
    server_addr.sin_addr.s_addr = INADDR_ANY;
    
    bind(sock, (struct sockaddr *)&server_addr, sizeof(server_addr));
    
    std::cout << "Сервер запущен и ожидает запросов." << std::endl;

    while (true) {
        struct sockaddr_in client_addr;
        socklen_t addr_len = sizeof(client_addr);
        char buffer[256];
        
        recvfrom(sock, buffer, sizeof(buffer), 0, (struct sockaddr *)&client_addr, &addr_len);
        
        double a, b;
        int n;
        sscanf(buffer, "INTEGRATE %lf %lf %d", &a, &b, &n);

        // Вычисление интеграла на заданном отрезке
        double result = compute_integral(function, a, b, n);

        // Отправка результата обратно мастеру
        snprintf(buffer, sizeof(buffer), "%f", result);
        sendto(sock, buffer, strlen(buffer), 0, (struct sockaddr *)&client_addr, addr_len);
    }

    close(sock);
    return 0;
}

double function(double x) {
    return x * x; // функция для интегрирования
}

double compute_integral(double (*f)(double), double a, double b, int n) {
    double h = (b - a) / n;
    double sum = 0.0;
    for (int i = 0; i < n; ++i) {
        sum += f(a + i * h) * h;
    }
    return sum;
}
