# Finances API 💰

API RESTful desarrollada en **Go** para la gestión de finanzas personales. Permite a los usuarios registrar ingresos y gastos, generar reportes mensuales y anuales, y calcular métricas financieras como diezmos, ofrendas y liquidaciones.

El proyecto implementa una **Arquitectura Limpia (Clean Architecture)** modular, separando responsabilidades para garantizar escalabilidad y fácil mantenimiento.

## 🚀 Tecnologías

* **Lenguaje:** Go (Golang)
* **Framework Web:** Fiber v2
* **Base de Datos:** MongoDB (Driver oficial)
* **Autenticación:** JWT (JSON Web Tokens) & Bcrypt
* **Infraestructura:** Docker (opcional)

## 📂 Estructura del Proyecto

El código sigue un flujo de datos unidireccional:
`Request HTTP` -> `Handler` -> `Service` -> `Repository` -> `MongoDB`

| Carpeta | Responsabilidad |
| :--- | :--- |
| **`config/`** | Configuración de la base de datos y conexión a MongoDB (`ConnectDB`). |
| **`handlers/`** | **Capa de Presentación:** Recibe peticiones HTTP, valida inputs y responde con JSON. No contiene lógica de negocio. |
| **`services/`** | **Capa de Lógica de Negocio:** Realiza cálculos (totales, porcentajes), validaciones complejas y orquesta transacciones. |
| **`repositories/`** | **Capa de Acceso a Datos:** Único punto de contacto con MongoDB. Ejecuta queries (Insert, Find, Update, Delete). |
| **`models/`** | Definición de estructuras de datos (`structs`) y etiquetas BSON/JSON (User, Report). |
| **`routes/`** | **Wiring (Cableado):** Configura las rutas e inyecta las dependencias (`Repo` -> `Service` -> `Handler`). |
| **`middleware/`** | Interceptores para proteger rutas, validando el token JWT. |
| **`main.go`** | Punto de entrada. Inicia la DB y levanta el servidor Fiber. |
