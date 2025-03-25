# 75-Day Challenge Tracker App Plan

## Overview

A comprehensive web and mobile application for tracking personalized 75-day challenges, inspired by the "Hard 75" concept but allowing for complete customization of challenge components, tasks, and tracking metrics.

## Tech Stack

### Frontend Web
- **Framework**: TanStack Start (React-based)
- **State Management**: 
  - Server State: TanStack Query
  - Client State: Zustand
- **Forms**: React Hook Form + Zod
- **UI**: Shadcn UI + Tailwind CSS
- **Data Visualization**: Recharts

### Mobile
- **Framework**: React Native
- **State Management**: Zustand (shared with web)
- **Forms**: React Hook Form
- **UI**: TBD (potentially NativeBase)

### Backend
- **Language**: Go
- **Database**: Neon (PostgreSQL)
- **Authentication**: Clerk

## Core Features

### 1. Customizable Challenge Structure
- Users can create custom challenge templates
- Ability to add, edit, and remove sections (e.g., Morning Routine, Physical Activity)
- Ability to add, edit, and remove tasks within each section
- Option to use predefined templates as starting points
- Default template based on provided markdown

### 2. Daily Tracking
- Digital implementation of the daily tracking template
- Progress photos upload and storage
- Workout logging with type, duration, and intensity
- Water intake tracking with reminders
- Meal logging and sugar consumption monitoring
- Reading and meditation tracking

### 3. Progress Visualization
- Streak tracking with visual calendar
- Measurement graphs over time
- Energy and mood trends
- Completion percentage by task and section

### 4. Accountability Features
- Failure detection and reset functionality
- Strike system for specific rules (e.g., sugar consumption)
- Daily notifications and reminders
- Weekly summary reports

## Database Schema

### Tables

#### Users
- `id`: UUID (primary key)
- `clerk_id`: String (reference to Clerk user)
- `email`: String
- `name`: String
- `created_at`: Timestamp
- `updated_at`: Timestamp

#### Challenges
- `id`: UUID (primary key)
- `user_id`: UUID (foreign key to Users)
- `name`: String
- `description`: String
- `start_date`: Date
- `end_date`: Date
- `current_day`: Integer
- `status`: Enum (active, completed, failed)
- `created_at`: Timestamp
- `updated_at`: Timestamp

#### Sections
- `id`: UUID (primary key)
- `challenge_id`: UUID (foreign key to Challenges)
- `name`: String
- `description`: String
- `order`: Integer
- `created_at`: Timestamp
- `updated_at`: Timestamp

#### Tasks
- `id`: UUID (primary key)
- `section_id`: UUID (foreign key to Sections)
- `name`: String
- `description`: String
- `task_type`: Enum (boolean, number, text, select, etc.)
- `required`: Boolean
- `restart_on_fail`: Boolean
- `strikes_enabled`: Boolean
- `strikes_limit`: Integer
- `order`: Integer
- `created_at`: Timestamp
- `updated_at`: Timestamp

#### DailyEntries
- `id`: UUID (primary key)
- `challenge_id`: UUID (foreign key to Challenges)
- `day_number`: Integer
- `date`: Date
- `completed`: Boolean
- `notes`: Text
- `progress_photo_url`: String
- `energy_level`: Integer
- `mood_level`: Integer
- `created_at`: Timestamp
- `updated_at`: Timestamp

#### TaskEntries
- `id`: UUID (primary key)
- `daily_entry_id`: UUID (foreign key to DailyEntries)
- `task_id`: UUID (foreign key to Tasks)
- `completed`: Boolean
- `value`: JSON (stores different data types based on task_type)
- `notes`: Text
- `created_at`: Timestamp
- `updated_at`: Timestamp

#### Measurements
- `id`: UUID (primary key)
- `challenge_id`: UUID (foreign key to Challenges)
- `day_number`: Integer
- `date`: Date
- `weight`: Decimal
- `chest`: Decimal
- `waist`: Decimal
- `hips`: Decimal
- `arms`: Decimal
- `thighs`: Decimal
- `created_at`: Timestamp
- `updated_at`: Timestamp

## API Endpoints

### Authentication
- Handled by Clerk

### Challenges
- `GET /api/challenges` - List user's challenges
- `POST /api/challenges` - Create a new challenge
- `GET /api/challenges/:id` - Get challenge details
- `PUT /api/challenges/:id` - Update challenge
- `DELETE /api/challenges/:id` - Delete challenge
- `POST /api/challenges/:id/reset` - Reset challenge to day 1
- `GET /api/challenges/:id/progress` - Get challenge progress

### Sections
- `GET /api/challenges/:id/sections` - List sections for a challenge
- `POST /api/challenges/:id/sections` - Create a new section
- `PUT /api/sections/:id` - Update section
- `DELETE /api/sections/:id` - Delete section
- `PUT /api/sections/:id/order` - Reorder section

### Tasks
- `GET /api/sections/:id/tasks` - List tasks for a section
- `POST /api/sections/:id/tasks` - Create a new task
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task
- `PUT /api/tasks/:id/order` - Reorder task

### Daily Entries
- `GET /api/challenges/:id/entries` - List entries for a challenge
- `GET /api/challenges/:id/entries/:day` - Get entry for a specific day
- `POST /api/challenges/:id/entries` - Create/update entry for today
- `PUT /api/challenges/:id/entries/:day` - Update entry for a specific day

### Measurements
- `GET /api/challenges/:id/measurements` - List measurements for a challenge
- `POST /api/challenges/:id/measurements` - Add new measurement
- `PUT /api/measurements/:id` - Update measurement

## UI Wireframes (Conceptual)

### Web App Screens
1. Dashboard
2. Challenge Creation/Configuration
3. Daily Entry Form
4. Progress Tracking
5. Settings

### Mobile App Screens
1. Today View
2. Challenge Progress
3. Daily Entry Form
4. History
5. Profile/Settings

## Implementation Plan

### Phase 1: Foundation
- Set up project structure (web and backend)
- Implement authentication with Clerk
- Create database schema
- Implement basic API endpoints
- Build challenge configuration UI

### Phase 2: Core Functionality
- Implement daily tracking interface
- Build customizable sections and tasks
- Create progress visualization components
- Implement failure rules and streak tracking

### Phase 3: Mobile App
- Set up React Native project
- Implement core screens
- Ensure sync with web app
- Test cross-platform functionality

### Phase 4: Refinement
- Add notifications and reminders
- Implement data export/import
- Polish UI/UX
- Performance optimization

## Feature Details: Customizable Challenge Structure

The cornerstone of this app is the ability for users to fully customize their challenge structure. Here's how it will work:

### Template System
- **Default Templates**: Provide pre-configured templates including the "75-Day Challenge" from the markdown
- **Custom Templates**: Allow users to create their own templates from scratch
- **Clone & Modify**: Enable users to clone existing templates and customize them

### Section Management
- Sections represent groupings of related tasks (e.g., "Morning Routine", "Physical Activity")
- Users can:
  - Add new sections
  - Edit section names and descriptions
  - Reorder sections via drag-and-drop
  - Delete sections

### Task Management
- Tasks are the individual items to be completed (e.g., "Take progress photo", "Complete workout")
- Different task types:
  - **Checkbox**: Simple complete/incomplete tasks
  - **Numeric**: Tasks with number values (e.g., water intake in ml)
  - **Duration**: Time-based tasks (e.g., workout duration)
  - **Rating**: Scale-based tasks (e.g., mood 1-10)
  - **Text**: Free text entry tasks (e.g., reflection writing)
  - **Photo**: Photo upload tasks
  
- Task properties:
  - Name
  - Description
  - Type (from above)
  - Required (whether the task is mandatory)
  - Restart rule (whether failing this task resets progress)
  - Strike system (accumulate strikes before resetting)
  - Default value/options

### Default Challenge Configuration

Based on the provided markdown template, the default challenge will include:

1. **Morning Routine Section**
   - Take progress photo (Checkbox/Photo)
   - Complete 20-minute meditation (Checkbox/Duration)
   - Avoid social media before 11 AM (Checkbox)
   - Plan meals for tomorrow (Checkbox)

2. **Physical Activity Section**
   - Complete workout/active rest (Checkbox)
   - Activity type (Select: Strength/Cardio/Active Rest)
   - Duration (Number, minutes)
   - Intensity (Scale 1-10)
   - Notes (Text)

3. **Nutrition & Hydration Section**
   - Water intake tracking (3 numeric entries for morning/afternoon/evening)
   - No added sugar consumed (Checkbox)
   - Logged all meals in food diary (Checkbox)
   - Stopped eating by 7 PM (Checkbox)

4. **Mental Development Section**
   - Read non-fiction (Checkbox)
   - Book title (Text)
   - Pages read (Number)
   - Complete task journaling (Checkbox)

5. **Evening Reflection Section**
   - Energy level (Scale 1-10)
   - Mood (Scale 1-10)
   - Daily reflection writing (Text)

6. **Weekly Measurements Section**
   - Weight (Number)
   - Body measurements (Multiple number fields)

7. **Progress Tracker Section**
   - Current streak (Auto-calculated)
   - Sugar strikes (Auto-calculated)
   - Overall progress (Auto-calculated)

This system will provide flexibility while maintaining the structure needed for effective challenge tracking.
