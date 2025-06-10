import { uiComponents } from './uiComponents.js';
import { formHandlers } from './formHandlers.js';
import { postService } from './postService.js';

const renderAndInit = async () => {
    const hash = window.location.hash || '#login';
    const app = document.getElementById('app');
    if (!app) {
        console.error('App container not found');
        return;
    }

    if (hash === '#signup') {
        app.innerHTML = uiComponents.renderSignup();
        setTimeout(() => formHandlers.initSignupForm(), 10);
    } else if (hash === '#login') {
        app.innerHTML = uiComponents.renderLogin();
        setTimeout(() => formHandlers.initLoginForm(), 10);
    } else if (hash === '#home') {
        try {
            const categories = await postService.getCategories();
            app.innerHTML = uiComponents.renderHome(categories);
            setTimeout(() => {
                formHandlers.initHomeHandlers();
                formHandlers.renderPosts(0); // Load all posts initially
            }, 10);
        } catch (error) {
            console.error('Error loading home page:', error);
            app.innerHTML = '<p>Error loading posts.</p>';
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