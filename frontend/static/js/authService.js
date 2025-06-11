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
        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
                credentials: 'include'
            });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Signup failed');
            }
            this.token = data.token;
            localStorage.setItem('token', this.token);
            this.userId = data.user.ID;
            this.nickname = data.user.nickname;
            return data.user;
        } catch (error) {
            throw error;
        }
    }

    async login(credentials) {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(credentials),
                credentials: 'include'
            });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Login failed');
            }
            this.token = data.token;
            localStorage.setItem('token', this.token);
            this.userId = data.user.id;
            this.nickname = data.user.nickname;
            return data.user;
        } catch (error) {
            throw error;
        }
    }

    async logout() {
        try {
            const response = await fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                credentials: 'include'
            });
            if (!response.ok) {
                throw new Error('Logout failed');
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            localStorage.removeItem('token');
            this.token = null;
            this.userId = null;
            this.nickname = '';
        }
    }

    isAuthenticated() {
        return !!this.token;
    }

    getUserInfo() {
        return {
            userId: this.userId,
            nickname: this.nickname
        };
    }

    getToken() {
        return this.token;
    }
}

const authService = new AuthService();
export default authService;