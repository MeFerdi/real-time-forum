import { uiComponents } from './uiComponents.js';
import formHandlers  from './formHandlers.js';
import postService from './postService.js';
import authService from './authService.js';

const checkSession = async () => {
    const token = authService.getToken();
    if (!token) return false;
    try {
        // Try a lightweight protected endpoint
        await postService.getCategories();
        return true;
    } catch {
        authService.logout(); // Clear token if invalid
        return false;
    }
};

// Add this helper to get current user info
const getCurrentUser = async () => {
    try {
        // Adjust the endpoint if needed
        const res = await fetch('/api/auth/me', {
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });
        if (!res.ok) throw new Error('Not authenticated');
        return await res.json();
    } catch {
        return {};
    }
};

const renderAndInit = async () => {
    let hash = window.location.hash || '#login';
    const app = document.getElementById('app');
    if (!app) {
        console.error('App container not found');
        return;
    }

    // Session check before rendering home
    if (hash === '#home') {
        const valid = await checkSession();
        if (!valid) {
            window.location.hash = '#login';
            return;
        }
    }

    if (hash === '#signup') {
        app.innerHTML = uiComponents.renderSignup();
        setTimeout(() => formHandlers.initSignupForm(), 10);
    } else if (hash === '#login') {
        app.innerHTML = uiComponents.renderLogin();
        setTimeout(() => formHandlers.initLoginForm(), 10);
    } else if (hash === '#home') {
        try {
            const [categories, user] = await Promise.all([
                postService.getCategories(),
                getCurrentUser()
            ]);
            console.log('Fetched categories:', categories);
            if (!categories || categories.length === 0) {
                console.warn('No categories fetched');
                app.innerHTML = uiComponents.renderHome([], user) + '<p>No categories available. Please try again later.</p>';
            } else {
                app.innerHTML = uiComponents.renderHome(categories, user);
            }
            setTimeout(() => {
                formHandlers.initHomeHandlers();
                formHandlers.renderPosts(0);
            }, 10);
        } catch (error) {
            console.error('Error loading home page:', error);
            app.innerHTML = uiComponents.renderHome([]) + '<p>Error loading categories. Please try again later.</p>';
        }
    } else {
        app.innerHTML = uiComponents.renderLogin();
        setTimeout(() => formHandlers.initLoginForm(), 10);
    }
};

window.addEventListener('hashchange', renderAndInit);
document.addEventListener('DOMContentLoaded', () => {
    renderAndInit();
});
export { renderAndInit };