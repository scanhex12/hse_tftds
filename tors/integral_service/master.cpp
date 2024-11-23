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
#include <queue>
#include <sys/epoll.h>
#include <unordered_set>
#include <vector>
#include <string>
#include <future>
#include <atomic>
#include <thread>
#include <functional>
#include <unordered_map>
#include <sys/epoll.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#include <fcntl.h>
#include <iostream>
#include <memory>
#include <stdexcept>
#include <cstring>
#include <algorithm> 

#define DISCOVERY_PORT 9002
#define WORK_PORT 9001
#define TIMEOUT_SEC 3
#define MAX_EVENTS 10

std::mutex mutex;
std::map<std::pair<int, std::string>, int> server_ports; 

void broadcast_search(int sock, struct sockaddr_in &broadcast_addr);
void receive_server_responses(int sock);
double integrate(double a, double b, int n, const std::vector<std::pair<int, std::string>> &servers);

class TaskPool {
public:
    struct Task {
        double a;
        double segment_length;
        int segment_size;
        size_t task_id;
    };

    explicit TaskPool(const std::vector<std::pair<int, std::string>>& servers, int timeout_ms = 15)
        : servers_(servers), timeout_ms_(timeout_ms) {}

    ~TaskPool() {
        for (int fd : server_fds_) {
            close(fd);
        }
    }

    std::future<double> Submit(Task task) {
        auto promise = std::make_shared<std::promise<double>>();
        std::future<double> result = promise->get_future();
        {
            tasks_.emplace_back(task, promise);
        }
        return result;
    }

    void Run() {
        processTasks();
    }

private:
    std::vector<std::pair<int, std::string>> servers_;
    std::vector<int> server_fds_;
    std::vector<bool> is_not_ok_server_;

    int timeout_ms_;
    std::vector<std::pair<Task, std::shared_ptr<std::promise<double>>>> tasks_;

    std::unordered_map<int, std::shared_ptr<std::promise<double>>> promises_;
    std::unordered_map<int, Task> current_cycle_tasks;
    
    std::unordered_map<int, int> sock_to_server_ind;
    std::vector<int> sockets_ready_to_process_new;
    std::unordered_set<int> processed_fds;

    int createSocket(const std::string& host, int port) {
        int sock = socket(AF_INET, SOCK_STREAM, 0);
        if (sock == -1) {
            std::cerr << "Socket creation failed: " << strerror(errno) << std::endl;
            return -1;
        }

        struct sockaddr_in server_addr;
        memset(&server_addr, 0, sizeof(server_addr));
        server_addr.sin_family = AF_INET;
        server_addr.sin_port = htons(port);
        if (inet_pton(AF_INET, host.c_str(), &server_addr.sin_addr) <= 0) {
            std::cerr << "Invalid address: " << strerror(errno) << std::endl;
            close(sock);
            return -1;
        }

        if (connect(sock, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
            std::cerr << "Connection failed: " << strerror(errno) << std::endl;
            close(sock);
            return -1;
        }

        return sock;
    }

    int GetNumUprocessedResponses() {
        int sum = 0;
        for (auto elem : server_fds_) {
            if (is_not_ok_server_[sock_to_server_ind[elem]]) {
                continue;
            }
            if (elem != -1) {
                std::cout << "Socket " << elem << " is not processed, but task are ended" << std::endl;
                sum++;
            }
        }
        return sum;
    }

    void processTasks() {
        int epoll_fd = epoll_create1(0);
        if (epoll_fd == -1) {
            std::cerr << "Failed to create epoll: " << strerror(errno) << std::endl;
            return;
        }

        std::vector<struct epoll_event> events(servers_.size());
        server_fds_ = std::vector<int>(servers_.size(), -1);
        is_not_ok_server_ = std::vector<bool>(servers_.size(), false);

        bool is_first_epoch = true;
        
        while (!tasks_.empty() || GetNumUprocessedResponses() != 0) {

            //
            // PRODUCER PART
            //

            bool end_stady = tasks_.empty() && (GetNumUprocessedResponses() != 0);
            std::vector<int> server_indices;
            if (is_first_epoch) {
                for (int i = 0; i < servers_.size(); ++i) {
                    if (is_not_ok_server_[i]) {
                        continue;
                    }

                    server_indices.push_back(i);
                }
            } else {
                for (auto sock : sockets_ready_to_process_new) {
                    if (is_not_ok_server_[sock_to_server_ind[sock]]) {
                        continue;
                    }

                    server_indices.push_back(sock_to_server_ind[sock]);
                }
                sockets_ready_to_process_new.clear();
            }
            if (end_stady) {
                server_indices.clear();
            }
            is_first_epoch = false;
            for (auto server_ind: server_indices) {
                std::cout << "SERVER " << server_ind << " IS READY TO PROCESS NEW TASKS" << std::endl;
                auto server = servers_[server_ind];
                int server_fd = createSocket(server.second, server.first);
                if (server_fd == -1) {
                    continue;
                }

                std::cout << "Server " << server_ind << " is response for socket " << server_fd << std::endl;

                server_fds_[server_ind] = server_fd;
                sock_to_server_ind[server_fd] = server_ind;
                struct epoll_event ev;
                ev.events = EPOLLIN | EPOLLET;
                ev.data.fd = server_fd;

                if (epoll_ctl(epoll_fd, EPOLL_CTL_ADD, server_fd, &ev) == -1) {
                    std::cerr << "Failed to add fd to epoll: " << strerror(errno) << std::endl;
                    close(server_fd);
                    continue;
                }
            }

            std::cout << "START EPOCH OF SENDING TASKS" << std::endl;
            for (int i : server_indices) {
                std::cout << "Sending task " << i << std::endl;
                sendTask(server_fds_[i]);
            }

            //
            // WAIT PART
            //


            int num_events = epoll_wait(epoll_fd, events.data(), events.size(), -1);
            if (num_events < 0) {
                std::cerr << "epoll_wait failed: " << strerror(errno) << std::endl;
                return;
            }

            if (num_events == 0) {
                std::cout << "epoll_wait timeout expired, no events." << std::endl;
                continue;
            }

            std::cout << "Events on epoll " << num_events << std::endl; 

            //
            // CONSUMER PART
            //

            for (int i = 0; i < num_events; ++i) {
                int fd = events[i].data.fd;
                if (processed_fds.count(fd)) {
                    continue;
                }
                processed_fds.insert(fd);
                std::cout << "Event on epoll on fd " << fd << std::endl; 
                sockets_ready_to_process_new.push_back(fd);

                if (events[i].events & EPOLLIN) {
                    std::cout << "Data is ready to read from fd " << fd << std::endl;
                    try {
                        receiveResponse(fd);
                    } catch (const std::exception& e) {
                        std::cerr << "Error receiving response: " << e.what() << std::endl;
                        close(fd);
                        epoll_ctl(epoll_fd, EPOLL_CTL_DEL, fd, nullptr);
                    }
                } else {
                    std::cout << "Unhandled event on fd " << fd << std::endl;
                }
            }
        }

        close(epoll_fd);
    }

    void sendTask(int fd) {
        Task task;
        std::shared_ptr<std::promise<double>> promise;

        {
            if (tasks_.empty()) return;

            task = tasks_.front().first;
            promise = tasks_.front().second;
            current_cycle_tasks[fd] = task;
            tasks_.erase(tasks_.begin());
        }

        if (!promise) {
            std::cerr << "Promise is null. Task cannot be processed." << std::endl;
            return;
        }

        std::cout << "Task " << task.task_id << " belongs to " << fd << ' ' <<sock_to_server_ind[fd] << std::endl;

        promises_[fd] = promise;

        double segment_start = task.a + task.task_id * task.segment_length;
        double segment_end = segment_start + task.segment_length;

        char message[256];
        snprintf(message, sizeof(message), "INTEGRATE %f %f %d", segment_start, segment_end, task.segment_size);

        std::cout << "PACKET TO SEND: " << message << std::endl;

        ssize_t sent = send(fd, message, strlen(message), 0);
        if (sent == -1) {
            std::cout << "Failed to send task to server: " << strerror(errno) << ". Reshedule" << std::endl;

            tasks_.push_back({current_cycle_tasks[fd], promises_[fd]});
            deleteServerByFd(fd);
            return;
        }

    }

    void deleteServerByFd(int fd) {
        is_not_ok_server_[sock_to_server_ind[fd]] = true;
    }

    void receiveResponse(int fd) {
        std::cout << "Start receiveResponse " << fd << std::endl;
        char buffer[256];
        ssize_t bytes_received = recv(fd, buffer, sizeof(buffer), 0);
        if (bytes_received <= 0) {
            std::cout << "Server disconnected or error occurred. Reshedule" << std::endl;
            if (!promises_[fd]) {
                throw std::runtime_error("Can not set null promise " + std::to_string(fd));
            }
            tasks_.push_back({current_cycle_tasks[fd], promises_[fd]});
            deleteServerByFd(fd);
            return;
        }

        buffer[bytes_received] = '\0';

        double result = atof(buffer);

        auto it = promises_.find(fd);
        if (it != promises_.end()) {
            it->second->set_value(result);
        } else {
            std::cerr << "No promise found for the server fd." << std::endl;
        }
        std::cout << "receiveResponse " << fd << ' ' << result << std::endl;
        server_fds_[sock_to_server_ind[fd]] = -1;
    }
};

int main() {
    std::this_thread::sleep_for(std::chrono::seconds(1));
    std::cout << "START MASTER SERVER" << std::endl;

    int sock = socket(AF_INET, SOCK_DGRAM, 0);
    struct sockaddr_in broadcast_addr;
    broadcast_addr.sin_family = AF_INET;
    broadcast_addr.sin_port = htons(DISCOVERY_PORT);
    broadcast_addr.sin_addr.s_addr = INADDR_BROADCAST;

    int broadcast_enable = 1;
    setsockopt(sock, SOL_SOCKET, SO_BROADCAST, &broadcast_enable, sizeof(broadcast_enable));

    broadcast_search(sock, broadcast_addr);
    
    receive_server_responses(sock);
    
    double a = 0.0, b = 10.0;
    int n = 1000;
    std::vector<std::pair<int, std::string>> servers;
    for (const auto &entry : server_ports) {
        servers.push_back(std::pair<int, std::string>{entry.first.first, entry.first.second});
    }
    
    if (servers.empty()) {
        std::cerr << "No available servers." << std::endl;
        close(sock);
        return -1;
    }
    std::cout << "Start to integrate" << std::endl;
    double result = integrate(a, b, n, servers);
    std::cout << "Result: " << result << std::endl;
    
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
    
    while (true) {
        FD_ZERO(&readfds);
        FD_SET(sock, &readfds);
        
        if (auto it = select(sock + 1, &readfds, NULL, NULL, &timeout); it > 0) {
            struct sockaddr_in server_addr;
            socklen_t addr_len = sizeof(server_addr);
            char buffer[256];
            
            recvfrom(sock, buffer, sizeof(buffer), 0, (struct sockaddr *)&server_addr, &addr_len);
            
            int work_port;
            sscanf(buffer, "MAIN_PORT %d", &work_port);
            std::string host = inet_ntoa(server_addr.sin_addr);

            std::cout << "Server is found: " << host << " port " << work_port << std::endl;

            std::lock_guard<std::mutex> lock(mutex);
            std::cout << "under lock" << std::endl;
            server_ports[{work_port, host}] = work_port;
            std::cout << "out under lock" << std::endl;
        } else {
            break;
        }
    }
}

double integrate(double a, double b, int n, const std::vector<std::pair<int, std::string>> &servers) {
    double total_result = 0.0;
    int segment_size = n / (servers.size() * 2);
    double segment_length = (b - a) / (servers.size() * 2.0);
    std::cout << "SEG LEN " << segment_length << std::endl;

    TaskPool pool(servers);
    std::vector<std::future<double>> futures;

    for (size_t i = 0; i < 2 * servers.size(); ++i) {
        futures.push_back(std::move(pool.Submit(TaskPool::Task{
            .a = a,
            .segment_length = segment_length,
            .segment_size = segment_size,
            .task_id = i,
        })));
    }
    pool.Run();

    for (size_t i = 0; i < 2 * servers.size(); ++i) {
        std::cout << "wait for future" << std::endl;
        total_result += futures[i].get();
    }

    return total_result;
}
