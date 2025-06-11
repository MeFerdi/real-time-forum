import authService from './authService.js';
class PostService {
    // postService.js (update createPost)
async createPost(formData) {
    const token = authService.getToken();
    if (!token) {
        throw new Error('No authentication token found. Please log in.');
    }
    console.log('Sending token:', token);
    try {
        const response = await fetch('/api/posts/create', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData,
            credentials: 'include'
        });
        if (!response.ok) {
            let errorMsg = `Failed to create post (HTTP ${response.status})`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.message || errorMsg;
            } catch {
                errorMsg = await response.text();
            }
            throw new Error(errorMsg);
        }
        return await response.json();
    } catch (error) {
        console.error('Create post error:', error.message);
        throw error;
    }
}

    async getPosts() {
        const response = await fetch('/api/posts', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });
        if (!response.ok) {
            throw new Error('Failed to load posts');
        }
        const data = await response.json();
        return data.posts || [];
    }

   async getCategories() {
    const response = await fetch('/api/categories', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authService.getToken()}`
        }
    });
    if (!response.ok) {
        throw new Error('Failed to load categories');
    }
    const data = await response.json();
    return data.categories;
}

    async addComment(postId, content) {
    const response = await fetch(`/api/posts/comments/${postId}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authService.getToken()}`
        },
        body: JSON.stringify({ content })
    });
    if (!response.ok) {
        let errorMsg = 'Failed to add comment';
        try {
            const error = await response.json();
            errorMsg = error.message || errorMsg;
        } catch {
            try {
                errorMsg = await response.text();
            } catch {}
        }
        throw new Error(errorMsg);
    }
    return await response.json();
}

    async updateComment(commentId, content) {
        const response = await fetch(`/api/comments/${commentId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: JSON.stringify({ content })
        });
        if (!response.ok) {
            throw new Error('Failed to update comment');
        }
        return await response.json();
    }

    async deleteComment(commentId) {
        const response = await fetch(`/api/comments/${commentId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });
        if (!response.ok) {
            throw new Error('Failed to delete comment');
        }
    }

    async handleReaction(postId, reactionType) {
        const response = await fetch(`/api/posts/react/${postId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: JSON.stringify({ type: reactionType })
        });
        if (!response.ok) {
            throw new Error('Failed to update reaction');
        }
        return await response.json();
    }
}
const postService = new PostService();
export default postService;