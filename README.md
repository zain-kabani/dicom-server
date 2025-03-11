# DICOM Server

A Go-based HTTP server for handling DICOM files, with features for uploading, storing, and retrieving DICOM images and metadata.

## Features

- Upload DICOM files
- Automatic PNG preview generation
- Metadata extraction and storage
- PostgreSQL integration for file metadata
- Structured file storage system
- RESTful API endpoints for data retrieval

## Setup

1. Clone the repository:
```bash
git clone <your-repository-url>
cd <repository-name>
```

2. Set up environment variables:
```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export POSTGRES_DB=imagedb
```

3. Initialize the database and start the app:
```bash
make start
```

The server will start on `http://localhost:3333`.

## API Endpoints

### Upload DICOM File
```http
POST /api/v1/images
Content-Type: multipart/form-data

Form field: image=@path/to/dicom/file.dcm
```

Response:
```json
{
    "message": "file uploaded successfully",
    "filename": "file.dcm",
    "file_id": 123
}
```

### Get DICOM Metadata
```http
GET /api/v1/dicom/{id}/header?tag=PatientName
```

Where {id} is in the response of the upload

Response:
```json
{
    "tag": "PatientName",
    "value": "John Doe"
}
```

### Get DICOM Preview
```http
GET /api/v1/dicom/{id}/preview
```

Where {id} is in the response of the upload

Returns a PNG image representing the DICOM file.

## File Storage Structure

Files are stored in the following structure:
```
uploads/
├── staging/     # Temporary storage for uploads
└── final/       # Permanent storage
    └── {hash}/  # Unique directory for each DICOM file
        ├── original.dcm
        └── preview.png
```

### Project Structure
```
.
├── internal/
│   ├── api/          # HTTP handlers and server setup
│   ├── dcmutil/      # DICOM processing utilities
│   ├── models/       # Data models
│   └── storage/      # Database and file storage logic
├── sql/             # Database scripts
└── main.go         # Application entry point
```

### Database Management
```bash
# Reset tables
make db-truncate

# Drop tables
make db-drop
``` 
