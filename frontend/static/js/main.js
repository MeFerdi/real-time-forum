import { uiComponents } from './uiComponents.js';
import formHandlers from './formHandlers.js';
import postService from './postService.js';
import authService from './authService.js';

const checkSession = async () => {
    const token = authService.getToken();
    if (!token) return false;
    try {
        await postService.getCategories();
        return true;
    } catch {
        authService.logout(); // Clear invalid token
        return false;
    }
};

const getCurrentUser = async () => {
    try {
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
    const app = document.getElementById('app');
    if (!app) {
        console.error('App container not found');
        return;
    }

    // Check session first, regardless of hash
    const isAuthenticated = await checkSession();

    // If not authenticated, force login page
    if (!isAuthenticated) {
        window.location.hash = '#login';
        app.innerHTML = uiComponents.renderLogin();
        setTimeout(() => formHandlers.initLoginForm(), 10);
        return;
    }

    // If authenticated, proceed with hash-based rendering
    let hash = window.location.hash || '#home'; // Default to home for authenticated users

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
        // Fallback to home for authenticated users with invalid hash
        window.location.hash = '#home';
        renderAndInit(); // Re-run to render home
    }
};

window.addEventListener('hashchange', renderAndInit);
document.addEventListener('DOMContentLoaded', () => {
    // Clear hash on initial load to avoid cached #home
    if (!authService.getToken()) {
        window.location.hash = '';
    }
    renderAndInit();
});

export { renderAndInit };