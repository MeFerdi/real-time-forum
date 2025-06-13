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
        authService.logout();
        return false;
    }
};

const getCurrentUser = async () => {
    try {
        const res = await fetch('/api/auth/me', {
            headers: { 'Authorization': `Bearer ${authService.getToken()}` }
        });
        if (!res.ok) throw new Error();
        return await res.json();
    } catch {
        return {};
    }
};

const renderAndInit = async () => {
    const app = document.getElementById('app');
    if (!app) return;

    let hash = window.location.hash || '#home';

    if (hash === '#signup') {
        app.innerHTML = uiComponents.renderSignup();
        setTimeout(formHandlers.initSignupForm, 10);
        return;
    }
    if (hash === '#login') {
        app.innerHTML = uiComponents.renderLogin();
        setTimeout(formHandlers.initLoginForm, 10);
        return;
    }

    if (!(await checkSession())) {
        window.location.hash = '#login';
        app.innerHTML = uiComponents.renderLogin();
        setTimeout(formHandlers.initLoginForm, 10);
        return;
    }

    if (hash === '#home') {
        try {
            const [categories, user] = await Promise.all([
                postService.getCategories(),
                getCurrentUser()
            ]);
            app.innerHTML = uiComponents.renderHome(categories || [], user);
            setTimeout(() => {
                formHandlers.initHomeHandlers();
                formHandlers.renderPosts(0);
            }, 10);
        } catch {
            app.innerHTML = uiComponents.renderHome([]) + '<p>Error loading categories.</p>';
        }
        return;
    }

    window.location.hash = '#home';
    renderAndInit();
};

window.addEventListener('hashchange', renderAndInit);
document.addEventListener('DOMContentLoaded', () => {
    if (!authService.getToken()) window.location.hash = '';
    renderAndInit();
});

export { renderAndInit };