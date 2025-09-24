import { request } from "@/utils/request";

export interface User {
  id: number;
  username: string;
  email: string;
  display_name: string;
  role: "admin" | "user" | "viewer";
  status: "active" | "inactive" | "banned";
  avatar?: string;
  last_login_at?: string;
  last_login_ip?: string;
  login_count: number;
  failed_attempts: number;
  locked_until?: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  display_name?: string;
  role: "admin" | "user" | "viewer";
}

export interface UpdateUserRequest {
  display_name?: string;
  email?: string;
  role?: string;
  status?: string;
  avatar?: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface UserListResponse {
  list: User[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_page: number;
  };
}

class UserAPI {
  // 用户登录
  login(data: LoginRequest) {
    return request.post<LoginResponse>("/users/login", data);
  }

  // 用户退出
  logout() {
    return request.post("/users/logout");
  }

  // 获取当前用户信息
  getProfile() {
    return request.get<User>("/users/profile");
  }

  // 更新当前用户信息
  updateProfile(data: UpdateUserRequest) {
    return request.put("/users/profile", data);
  }

  // 修改密码
  changePassword(data: ChangePasswordRequest) {
    return request.post("/users/change-password", data);
  }

  // 创建用户（管理员）
  createUser(data: CreateUserRequest) {
    return request.post<User>("/users", data);
  }

  // 获取用户列表（管理员）
  listUsers(params?: {
    page?: number;
    page_size?: number;
    role?: string;
    status?: string;
    search?: string;
  }) {
    return request.get<UserListResponse>("/users", { params });
  }

  // 获取指定用户（管理员）
  getUser(id: number) {
    return request.get<User>(`/users/${id}`);
  }

  // 更新用户（管理员）
  updateUser(id: number, data: UpdateUserRequest) {
    return request.put(`/users/${id}`, data);
  }

  // 删除用户（管理员）
  deleteUser(id: number) {
    return request.delete(`/users/${id}`);
  }
}

export const userAPI = new UserAPI();
