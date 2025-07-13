**Item Type**: Application

**Identifier**: `PK=APP#<gtId>`,`SK=APP#<gtId>`

**Attributes**:
- `GSI_PK` String, value: "APP", Used for filtering
- `CreatedAt`: String, When application was created in ISO 8601 format
- `Name`: String, name of the applicant
- `Email`: String, email of the applicant 
- `Analysed`: Boolean, whether the application was analysed or not
- `PreScreeningStatus`: String, Current status of pre-screening
- `AdminRequestedDocuments`: List of Strings ([]string), Documents requested by the admin that applicants should upload/update
- `AiSummary:`: String, Gist of the CV by AI 
- `Score`: Number, Score of the CV according to https://www.migration.gv.at/en/types-of-immigration/permanent-immigration/skilled-workers-in-shortage-occupations/

**Relationships**: Linked to other items via PK (documents)

**Example item:**

JSON

```json
{
  "PK": "APP#a1b2c3d4-e5f6-7890-1234-567890abcdef"
  "SK": "APP#a1b2c3d4-e5f6-7890-1234-567890abcdef"
  "GSI_PK": "APP",
  "CreatedAt": "2025-04-09T12:30:00Z",
  "Name": "John Doe",
  "Email": "john.doe@example.com",
  "Analysed": false,
  "PreScreeningStatus": "Pending Review",
  "AdminRequestedDocuments": [
    "Resume",
    "Cover Letter",
    "Transcript",
    "References"
  ]
}
```


**Item Type**: Post

**Identifier**: PK=P#<postId>, SK=P#<postId>

**Attributes**:
GSI_PK: String, value: "POST#<language>" Used for filtering posts by language.
CreatedAt: String, When the post was created in ISO 8601 format.
Title: String, Title of the post.
Content: String, Content of the post.

**Relationships**:  Linked to other documents via PK 

Example Item (English version):

```json
{
  "PK": "POST#f9e8d7c6-b5a4-3210-fedc-ba9876543210",
  "SK": "POST#f9e8d7c6-b5a4-3210-fedc-ba9876543210",
  "GSI_PK": "POST#en",
  "CreatedAt": "2025-04-09T12:34:00Z",
  "Title": "Welcome to our Blog!",
  "Content": "This is the first post on our brand new blog. Stay tuned for more exciting content!"
}
```

**Item Type**: Document

**Identifier**: PK=<parent_entity_pk>, SK=<document_type>

**Attributes**:
`Name`: String, The original name of the uploaded document.
`ContentType`: String, The MIME type of the document (e.g., "application/pdf", "image/jpeg").

**Relationships**: Directly linked to a parent entity (Application, Post, etc.) through the PK.

**Example**

JSON 

```json
{
  "PK": "APP#uuid123",
  "SK": "DCV",
  "Name": "APP#uuid123-DCV.pdf",
  "ContentType": "application/pdf"
}
```