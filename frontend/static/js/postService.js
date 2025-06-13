import authService from './authService.js';

class PostService {
    async createPost(formData) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to create a post.');
        try {
            const categoryId = formData.get('category_id');
            if (categoryId) {
                formData.set('categories', categoryId);
                formData.delete('category_id');
            }
            const response = await fetch('/api/posts/create', {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${token}` },
                body: formData,
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to create post: ${response.statusText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Create post error:', error);
            throw error;
        }
    }

    async getPosts(categoryId) {
    const token = authService.getToken();
    if (!token) throw new Error('Please log in to view posts.');
    try {
        const url = categoryId && categoryId !== '0' ? `/api/posts?category_id=${categoryId}` : '/api/posts';
        const response = await fetch(url, {
            headers: { 'Authorization': `Bearer ${token}` },
            credentials: 'include'
        });
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.message || `Failed to fetch posts: ${response.statusText}`);
        }
        const data = await response.json();
        console.log('Posts response:', data);
        if (!data || typeof data !== 'object') {
            throw new Error('Invalid posts response: data is null or not an object');
        }
        const posts = Array.isArray(data.posts) ? data.posts : [];
        for (const post of posts) {
            // Always ensure comments is an array, never null/undefined
            try {
                const comments = await this.getComments(post.id);
                post.comments = Array.isArray(comments) ? comments : [];
            } catch (err) {
                console.error('Fetch comments error:', err);
                post.comments = [];
            }
            post.current_user_id = authService.getCurrentUserId();
            if (post.user) {
                post.author_nickname = post.user.nickname;
            }
        }
        return posts;
    } catch (error) {
        console.error('Fetch posts error:', error);
        throw error;
    }
}

    async getPostsByUserID(userID) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to view posts.');
        try {
            const response = await fetch(`/api/posts?userID=${userID}`, {
                headers: { 'Authorization': `Bearer ${token}` },
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to fetch user posts: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('User posts response:', data);
            if (!data || typeof data !== 'object') {
                throw new Error('Invalid user posts response: data is null or not an object');
            }
            const posts = Array.isArray(data.posts) ? data.posts : [];
            for (const post of posts) {
                post.comments = await this.getComments(post.id);
                post.current_user_id = authService.getCurrentUserId();
                if (post.user) {
                    post.author_nickname = post.user.nickname;
                }
            }
            return posts;
        } catch (error) {
            console.error('Fetch user posts error:', error);
            throw error;
        }
    }

    async getLikedPosts() {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to view liked posts.');
        try {
            const response = await fetch('/api/posts/liked', {
                headers: { 'Authorization': `Bearer ${token}` },
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to fetch liked posts: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Liked posts response:', data);
            if (!data || typeof data !== 'object') {
                throw new Error('Invalid liked posts response: data is null or not an object');
            }
            const posts = Array.isArray(data.posts) ? data.posts : [];
            for (const post of posts) {
                post.comments = await this.getComments(post.id);
                post.current_user_id = authService.getCurrentUserId();
                if (post.user) {
                    post.author_nickname = post.user.nickname;
                }
            }
            return posts;
        } catch (error) {
            console.error('Fetch liked posts error:', error);
            throw error;
        }
    }

    async getCategories() {
        try {
            const token = authService.getToken();
            const headers = token ? { 'Authorization': `Bearer ${token}` } : {};
            const response = await fetch('/api/categories', {
                headers,
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to fetch categories: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Categories response:', data);
            if (!data || typeof data !== 'object') {
                throw new Error('Invalid categories response: data is null or not an object');
            }
            return Array.isArray(data.Categories) ? data.Categories : Array.isArray(data) ? data : [];
        } catch (error) {
            console.error('Error fetching categories:', error);
            return [];
        }
    }

    async getMessages() {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to view messages.');
        try {
            const response = await fetch('/api/messages', {
                headers: { 'Authorization': `Bearer ${token}` },
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to fetch messages: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Messages response:', data);
            if (!data || typeof data !== 'object') {
                throw new Error('Invalid messages response: data is null or not an object');
            }
            return Array.isArray(data.Messages) ? data.Messages : [];
        } catch (error) {
            console.error('Fetch messages error:', error);
            throw error;
        }
    }

    async getMessage(messageId) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to view message.');
        try {
            const response = await fetch(`/api/messages/${messageId}`, {
                headers: { 'Authorization': `Bearer ${token}` },
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to fetch message: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Message response:', data);
            if (!data) {
                throw new Error('Invalid message response: data is null');
            }
            return data;
        } catch (error) {
            console.error('Fetch message error:', error);
            throw error;
        }
    }

    async getComments(postId) {
    const token = authService.getToken();
    if (!token) throw new Error('Please log in to view comments.');
    try {
        const response = await fetch(`/api/posts/comments?post_id=${postId}`, {
            headers: { 'Authorization': `Bearer ${token}` },
            credentials: 'include'
        });
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.message || `Failed to fetch comments: ${response.statusText}`);
        }
        const data = await response.json();
        console.log('Comments response:', data);
        if (Array.isArray(data)) return data;
        if (data && typeof data === 'object' && Array.isArray(data.Comments)) return data.Comments;
        return [];
    } catch (error) {
        console.error('Fetch comments error:', error);
        return [];
    }
}
    async addComment(postId, content) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to comment.');
        try {
            const response = await fetch(`/api/posts/comments/${postId}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ content }),
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to add comment: ${response.statusText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Add comment error:', error);
            throw error;
        }
    }

    async editComment(commentId, content) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to edit comment.');
        try {
            const response = await fetch(`/api/comments/${commentId}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ content }),
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to update comment: ${response.statusText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Edit comment error:', error);
            throw error;
        }
    }

    async deleteComment(commentId) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to delete comment.');
        try {
            const response = await fetch(`/api/comments/${commentId}`, {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` },
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to delete comment: ${response.statusText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Delete comment error:', error);
            throw error;
        }
    }

    async handleReaction(postId, reactionType) {
        const token = authService.getToken();
        if (!token) throw new Error('Please log in to react.');
        try {
            const response = await fetch(`/api/posts/react/${postId}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ type: reactionType }),
                credentials: 'include'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.message || `Failed to update reaction: ${response.statusText}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Handle reaction error:', error);
            throw error;
        }
    }
}

const postService = new PostService();
export default postService;