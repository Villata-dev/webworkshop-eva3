console.log("app.js cargado y ejecutando.");

// Aquí es donde irá la lógica principal para inicializar la app,
// registrar Web Components y manejar el flujo de la aplicación.

// Inicializar la aplicación cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
    console.log("Inicializando aplicación...");

    const authForm = document.querySelector('auth-form');
    const productsList = document.querySelector('editable-list');
    const productsSection = document.getElementById('products-section');

    // Almacenar el usuario actual
    let currentUser = null;

    // Verificación inicial de elementos
    console.log("Estado inicial de elementos:", {
        authForm: !!authForm,
        productsList: !!productsList,
        productsSection: !!productsSection
    });

    if (!authForm) {
        console.error("No se encontró el componente auth-form");
        return;
    }

    // Escuchar el evento auth-success del componente
    authForm.addEventListener('auth-success', (event) => {
        console.log("Evento auth-success recibido");
        console.log("Usuario autenticado:", event.detail.user);
        
        // Verificar que productsSection existe antes de usarlo
        if (!productsSection) {
            console.error("No se encontró la sección de productos");
            return;
        }

        // Mostrar la sección de productos
        productsSection.style.display = 'block';
        authForm.style.display = 'none';
        loadProducts();
    });

    // Manejar el evento de logout
    authForm.addEventListener('auth-logout', () => {
        productsSection.style.display = 'none';
        productsList.setData([]);
    });
    
    // Función para actualizar la UI según el usuario
    function updateUIForUser(user) {
        console.log("Actualizando UI para usuario:", user);
        
        // Obtener la referencia al productsSection nuevamente
        const productsSection = document.getElementById('products-section');
        if (!productsSection) {
            console.error("No se encontró la sección de productos para actualizar UI");
            return;
        }

        // Verificar que user y user.rol existen
        if (!user || !user.rol) {
            console.error("Datos de usuario inválidos:", user);
            return;
        }

        // Limpiar clases anteriores
        productsSection.classList.remove('admin-view', 'user-view');
        
        // Añadir la clase correspondiente
        const role = user.rol.toLowerCase();
        if (role === 'admin') {
            console.log("Aplicando vista de admin");
            productsSection.classList.add('admin-view');
        } else {
            console.log("Aplicando vista de usuario normal");
            productsSection.classList.add('user-view');
        }
        
        // Cargar la lista de productos después del login
        loadProducts();
    }

    // Función para cargar productos
    async function loadProducts() {
        console.log("Intentando cargar productos...");
        try {
            const response = await fetch('/api/v1/productos', {
                credentials: 'include'
            });
            
            if (!response.ok) {
                if (response.status === 401) {
                    handleSessionExpired();
                    return;
                }
                throw new Error('Error al cargar productos');
            }
            
            const data = await response.json();
            console.log("Productos recibidos:", data);
            
            if (productsList && data && Array.isArray(data.items)) {
                productsList.setData(data.items);
                console.log("Productos cargados en el componente");
            } else {
                console.error("Error: formato de datos inválido o componente no encontrado");
            }
        } catch (error) {
            console.error('Error al cargar productos:', error);
            mostrarMensaje(error.message, 'error');
        }
    }
    
    // Manejar eventos del listado de productos
    productsList.addEventListener('item-create', async (e) => {
        try {
            const response = await fetch('/api/v1/productos', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(e.detail.item),
                credentials: 'include' // Importante: incluir cookies
            });
            
            if (!response.ok) {
                if (response.status === 401) {
                    // Sesión expirada
                    handleSessionExpired();
                    return;
                }
                throw new Error('Error al crear producto');
            }
            
            await loadProducts(); // Recargar la lista
            mostrarMensaje('Producto creado exitosamente', 'success');
        } catch (error) {
            console.error('Error:', error);
            mostrarMensaje(error.message, 'error');
        }
    });
    
    productsList.addEventListener('item-delete', async (e) => {
        try {
            const response = await fetch(`/api/v1/productos/${e.detail.id}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (!response.ok) throw new Error('Error al eliminar producto');
            
            loadProducts();
            mostrarMensaje('Producto eliminado exitosamente', 'success');
        } catch (error) {
            mostrarMensaje(error.message, 'error');
        }
    });

    // Añadir el manejador de edición después del manejador de eliminación
    productsList.addEventListener('item-edit', async (e) => {
        try {
            const response = await fetch(`/api/v1/productos/${e.detail.item.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(e.detail.item),
                credentials: 'include' // Para enviar cookies de autenticación
            });
            
            if (!response.ok) throw new Error('Error al actualizar producto');
            
            loadProducts();
            mostrarMensaje('Producto actualizado exitosamente', 'success');
        } catch (error) {
            mostrarMensaje(error.message, 'error');
        }
    });

    // Función para manejar sesión expirada
    function handleSessionExpired() {
        currentUser = null;
        productsSection.style.display = 'none';
        authForm.style.display = 'block';
        mostrarMensaje('Sesión expirada. Por favor, inicie sesión nuevamente.', 'error');
    }

    // Función auxiliar para mostrar mensajes
    function mostrarMensaje(mensaje, tipo) {
        const mensajeEl = document.createElement('div');
        mensajeEl.className = `mensaje ${tipo}`;
        mensajeEl.textContent = mensaje;
        document.body.appendChild(mensajeEl);

        // Aplicar estilos
        Object.assign(mensajeEl.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '10px 20px',
            borderRadius: '4px',
            backgroundColor: tipo === 'success' ? '#4CAF50' : '#f44336',
            color: 'white',
            zIndex: '1000'
        });

        // Remover después de 3 segundos
        setTimeout(() => {
            mensajeEl.remove();
        }, 3000);
    }
});