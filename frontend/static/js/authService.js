class AuthService {
    constructor() {
        this.token = localStorage.getItem('token');
        this.userId = null;
        this.nickname = '';
    }

    async signup(userData) {
        const payload = {
            email: userData.email,
            nickname: userData.nickname,
            password: userData.password,
            first_name: userData.firstName,
            last_name: userData.lastName,
            age: userData.age,
            gender: userData.gender
        };
        const res = await fetch('/api/auth/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Signup failed');
        this.token = data.token;
        localStorage.setItem('token', this.token);
        this.userId = data.user.ID;
        this.nickname = data.user.nickname;
        return data.user;
    }

    async login(credentials) {
        const res = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                identifier: credentials.identifier,
                password: credentials.password
            })
        });
        const data = await res.json();
        if (!res.ok || !data.token) throw new Error(data.message || 'Login failed');
        localStorage.setItem('token', data.token);
        this.token = data.token;
        this.userId = data.user.ID;
        this.nickname = data.user.nickname;
        window.location.hash = '#home';
        return data.user;
    }

    getToken() {
        const token = localStorage.getItem('token');
        if (!token) {
            window.location.hash = '#login';
            return null;
        }
        return token;
    }

    logout() {
        const token = this.getToken();
        if (token) {
            fetch('/api/auth/logout', {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` }
            }).finally(() => {
                localStorage.removeItem('token');
                window.location.hash = '#login';
            });
        } else {
            window.location.hash = '#login';
        }
    }

    async getUser() {
        const token = this.getToken();
        if (!token) throw new Error('No valid token found');
        const res = await fetch('/api/auth/profile', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!res.ok) throw new Error('Failed to fetch user profile');
        const data = await res.json();
        this.userId = data.ID;
        this.nickname = data.nickname;
        return data;
    }
    getCurrentUserId() {
        return this.userId;
    }
}

const authService = new AuthService();
export default authService;