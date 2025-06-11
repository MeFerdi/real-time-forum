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
            body: JSON.stringify({
                identifier: credentials.identifier,
                password: credentials.password
            })
        });
        const data = await response.json();
        // Accept both camelCase and snake_case for expiresAt
        const expiresAt = data.expiresAt || data.expires_at;
        if (response.ok) {
            if (!data.token || !expiresAt) {
                console.error('Login response missing token or expiresAt:', data);
                throw new Error('Invalid login response');
            }
            localStorage.setItem('token', data.token);
            localStorage.setItem('expiresAt', expiresAt);
            console.log('Stored token:', data.token, 'Expires:', expiresAt);
            window.location.hash = '#home';
            return data.user;
        } else {
            throw new Error(data.message || 'Login failed');
        }
    } catch (error) {
        console.error('Login error:', error.message);
        throw error;
    }
}
    getToken() {
        const token = localStorage.getItem('token');
        const expiresAt = localStorage.getItem('expiresAt');
        if (!token || !expiresAt) {
            console.warn('No token or expiresAt found');
            return null;
        }
        const expiryDate = new Date(expiresAt);
        if (expiryDate < new Date()) {
            console.warn('Session expired at:', expiresAt);
            localStorage.removeItem('token');
            localStorage.removeItem('expiresAt');
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
            }).then(() => {
                localStorage.removeItem('token');
                localStorage.removeItem('expiresAt');
                window.location.hash = '#login';
            });
        } else {
            window.location.hash = '#login';
        }
    }
    async getUser() {
        const token = this.getToken();
        if (!token) {
            throw new Error('No valid token found');
        }
        try {
            const response = await fetch('/api/auth/profile', {
                method: 'GET',
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (!response.ok) {
                throw new Error('Failed to fetch user profile');
            }
            const data = await response.json();
            this.userId = data.ID;
            this.nickname = data.nickname;
            return data;
        } catch (error) {
            throw error;
        }
    }
}

const authService = new AuthService();

export default authService;