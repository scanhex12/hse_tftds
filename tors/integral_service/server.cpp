#include <chrono>
#include <iostream>
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <cstring>
#include <thread>

#define MAIN_PORT 9001
#define DISCOVERY_PORT 9002

double compute_integral(double (*f)(double), double a, double b, int n);
double function(double x);

void discovery_server();
void main_server();

int main() {
    std::thread discovery_thread(discovery_server);
    std::thread main_thread(main_server);

    discovery_thread.join();
    main_thread.join();

    return 0;
}

void discovery_server() {
    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    if (sock < 0) {
        perror("Discovery socket creation failed");
        exit(EXIT_FAILURE);
    }

    struct sockaddr_in server_addr;
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(DISCOVERY_PORT);
    server_addr.sin_addr.s_addr = INADDR_ANY;

    if (bind(sock, (struct sockaddr *)&server_addr, sizeof(server_addr)) < 0) {
        perror("Discovery socket bind failed");
        close(sock);
        exit(EXIT_FAILURE);
    }

    std::cout << "Discovery server is running on " << DISCOVERY_PORT << "." << std::endl;

    while (true) {
        struct sockaddr_in client_addr;
        socklen_t addr_len = sizeof(client_addr);
        char buffer[256];

        std::cout << "Start to getting request" << std::endl;

        recvfrom(sock, buffer, sizeof(buffer), 0, (struct sockaddr *)&client_addr, &addr_len);

        std::cout << "Start to process request" << std::endl;

        snprintf(buffer, sizeof(buffer), "MAIN_PORT %d", MAIN_PORT);
        sendto(sock, buffer, strlen(buffer), 0, (struct sockaddr *)&client_addr, addr_len);
    }

    close(sock);
}

double compute_integral(double a, double b, int n) {
    std::this_thread::sleep_for(std::chrono::seconds(10));
    double h = (b - a) / n;
    double sum = 0.0;
    for (int i = 0; i < n; ++i) {
        sum += (a + i * h) * (a + i * h) * h;
    }
    return sum;
}

void main_server() {
    int sock = socket(AF_INET, SOCK_STREAM, 0); 
    if (sock < 0) {
        perror("Main socket creation failed");
        exit(EXIT_FAILURE);
    }

    struct sockaddr_in server_addr{};
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(MAIN_PORT);
    server_addr.sin_addr.s_addr = INADDR_ANY;

    if (bind(sock, (struct sockaddr *)&server_addr, sizeof(server_addr)) < 0) {
        perror("Main server bind failed");
        close(sock);
        exit(EXIT_FAILURE);
    }

    if (listen(sock, 5) < 0) {
        perror("Listen failed");
        close(sock);
        exit(EXIT_FAILURE);
    }

    std::cout << "Server is listening on port " << MAIN_PORT << std::endl;

    while (true) {
        struct sockaddr_in client_addr{};
        socklen_t addr_len = sizeof(client_addr);
        
        int client_sock = accept(sock, (struct sockaddr *)&client_addr, &addr_len);
        if (client_sock < 0) {
            perror("Accept failed");
            continue;
        }

        std::cout << "Client connected." << std::endl;

        char buffer[256];
        
        ssize_t bytes_received = recv(client_sock, buffer, sizeof(buffer), 0);
        if (bytes_received < 0) {
            perror("recv failed");
            close(client_sock);
            continue;
        }

        std::cout << "Task received: " << buffer << std::endl;

        double a, b;
        int n;
        if (sscanf(buffer, "INTEGRATE %lf %lf %d", &a, &b, &n) == 3) {
            std::cout << "Task parsed: a=" << a << ", b=" << b << ", n=" << n << std::endl;
            double result = compute_integral(a, b, n); 
            std::cout << "Computing ended: " << result << std::endl;

            snprintf(buffer, sizeof(buffer), "%f", result);
            send(client_sock, buffer, strlen(buffer), 0);
            std::cout << "Result sent back: " << result << std::endl;
        } else {
            std::cout << "Failed to parse the task" << std::endl;
        }

        close(client_sock); 
    }

    close(sock);  
}
