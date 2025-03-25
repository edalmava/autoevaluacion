// web/static/js/auth-check.js
document.addEventListener('DOMContentLoaded', function() {
    // Verificar si hay un token
    const token = localStorage.getItem('token');
    
    if (!token) {
        // No hay token, redirigir a login
        window.location.href = '/login';
        return;
    }
    
    // Verificar si el token es válido
    fetch('/api/users/me', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Token inválido');
        }
        return response.json();
    })
    .then(user => {
        // Token válido, guardar usuario
        localStorage.setItem('user', JSON.stringify(user));
        
        // Verificar permisos según la página
        const path = window.location.pathname;
        
        if (path === '/admin' && user.role !== 'admin') {
            // Acceso denegado a admin para no-administradores
            window.location.href = '/evaluacion';
        }
    })
    .catch(error => {
        console.error('Error de autenticación:', error);
        // Eliminar token inválido
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        // Redirigir a login
        window.location.href = '/login';
    });
    
    // Configurar el header de Authorization para todas las solicitudes fetch
    const originalFetch = window.fetch;
    window.fetch = function(url, options = {}) {
        options.headers = options.headers || {};
        
        // Solo añadir el token a las solicitudes a nuestra API
        if (url.startsWith('/api/')) {
            options.headers['Authorization'] = `Bearer ${token}`;
        }
        
        return originalFetch(url, options);
    };
});