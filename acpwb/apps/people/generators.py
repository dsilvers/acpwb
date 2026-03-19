import random
import hashlib

FIRST_NAMES = [
    'James', 'Mary', 'John', 'Patricia', 'Robert', 'Jennifer', 'Michael', 'Linda',
    'William', 'Barbara', 'David', 'Elizabeth', 'Richard', 'Susan', 'Joseph', 'Jessica',
    'Thomas', 'Sarah', 'Charles', 'Karen', 'Christopher', 'Lisa', 'Daniel', 'Nancy',
    'Matthew', 'Betty', 'Anthony', 'Margaret', 'Mark', 'Sandra', 'Donald', 'Ashley',
    'Steven', 'Dorothy', 'Paul', 'Kimberly', 'Andrew', 'Emily', 'Joshua', 'Donna',
    'Kenneth', 'Michelle', 'Kevin', 'Carol', 'Brian', 'Amanda', 'George', 'Melissa',
    'Timothy', 'Deborah', 'Ronald', 'Stephanie', 'Edward', 'Rebecca', 'Jason', 'Sharon',
    'Jeffrey', 'Laura', 'Ryan', 'Cynthia', 'Jacob', 'Kathleen', 'Gary', 'Amy',
    'Nicholas', 'Angela', 'Eric', 'Shirley', 'Jonathan', 'Anna', 'Stephen', 'Brenda',
    'Larry', 'Pamela', 'Justin', 'Emma', 'Scott', 'Nicole', 'Brandon', 'Helen',
    'Benjamin', 'Samantha', 'Samuel', 'Katherine', 'Raymond', 'Christine', 'Gregory',
    'Debra', 'Frank', 'Rachel', 'Alexander', 'Carolyn', 'Patrick', 'Janet', 'Jack',
    'Catherine', 'Dennis', 'Maria', 'Jerry', 'Heather', 'Tyler', 'Diane', 'Aaron',
    'Julie', 'Jose', 'Joyce', 'Adam', 'Victoria', 'Nathan', 'Kelly', 'Henry',
    'Christina', 'Douglas', 'Lauren', 'Zachary', 'Joan', 'Peter', 'Evelyn', 'Kyle',
    'Olivia', 'Walter', 'Judith', 'Ethan', 'Megan', 'Jeremy', 'Cheryl', 'Harold',
    'Andrea', 'Terry', 'Hannah', 'Sean', 'Martha', 'Austin', 'Jacqueline', 'Carl',
    'Frances', 'Arthur', 'Gloria', 'Lawrence', 'Ann', 'Dylan', 'Teresa', 'Jesse',
    'Kathryn', 'Jordan', 'Sara', 'Bryan', 'Janice', 'Billy', 'Jean', 'Bruce',
    'Alice', 'Gabriel', 'Madison', 'Joe', 'Doris', 'Logan', 'Abigail', 'Alan',
    'Julia', 'Juan', 'Judy', 'Albert', 'Rose', 'Willie', 'Grace', 'Wayne',
    'Denise', 'Randy', 'Amber', 'Mason', 'Marilyn', 'Roy', 'Beverly', 'Ralph',
    'Danielle', 'Eugene', 'Theresa', 'Russell', 'Sophia', 'Bobby', 'Marie',
    # additional diverse names
    'Aisha', 'Priya', 'Wei', 'Mei', 'Fatima', 'Amara', 'Yuki', 'Leila',
    'Ananya', 'Zara', 'Nadia', 'Ingrid', 'Astrid', 'Brigitte', 'Colette',
    'Daphne', 'Elise', 'Fiona', 'Gwendolyn', 'Helena', 'Isadora', 'Josephine',
    'Kieran', 'Liam', 'Niall', 'Ronan', 'Cormac', 'Declan', 'Fergus', 'Seamus',
    'Arjun', 'Vikram', 'Rahul', 'Nikhil', 'Ravi', 'Sanjay', 'Deepak', 'Amit',
    'Omar', 'Tariq', 'Hassan', 'Khalid', 'Ibrahim', 'Yousef', 'Faisal', 'Nasser',
    'Diego', 'Mateo', 'Alejandro', 'Sebastian', 'Ricardo', 'Fernando', 'Carlos', 'Miguel',
    'Elena', 'Isabel', 'Valentina', 'Camila', 'Lucia', 'Sofia', 'Gabriela', 'Catalina',
    'Hiroshi', 'Kenji', 'Takeshi', 'Naoki', 'Shinji', 'Haruto', 'Ren', 'Yuto',
    'Akiko', 'Yumi', 'Keiko', 'Naomi', 'Haruki', 'Sakura', 'Aiko', 'Emi',
    'Chidi', 'Emeka', 'Obinna', 'Adebayo', 'Oluwaseun', 'Nnamdi', 'Chisom', 'Ifeoma',
    'Kofi', 'Kwame', 'Ama', 'Abena', 'Yaw', 'Efua', 'Nana', 'Akosua',
    'Andrei', 'Dmitri', 'Pavel', 'Nikolai', 'Alexei', 'Sergei', 'Mikhail', 'Boris',
    'Anastasia', 'Natasha', 'Irina', 'Olga', 'Tatiana', 'Svetlana', 'Ekaterina',
    'Henrik', 'Lars', 'Bjorn', 'Erik', 'Magnus', 'Sven', 'Gunnar', 'Torsten',
    'Annika', 'Sigrid', 'Freya', 'Solveig', 'Maja', 'Lena', 'Hanna', 'Karin',
    'Amir', 'Darius', 'Cyrus', 'Xerxes', 'Reza', 'Mehdi', 'Navid', 'Shervin',
    'Aroha', 'Tane', 'Hemi', 'Rangi', 'Wiremu', 'Aroha', 'Mere', 'Hine',
    'Cameron', 'Callum', 'Ewan', 'Angus', 'Hamish', 'Alistair', 'Iain', 'Rory',
    'Fionnuala', 'Siobhan', 'Aoife', 'Niamh', 'Caoimhe', 'Aisling', 'Ciara', 'Grainne',
]

LAST_NAMES = [
    'Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis',
    'Rodriguez', 'Martinez', 'Hernandez', 'Lopez', 'Gonzalez', 'Wilson', 'Anderson',
    'Thomas', 'Taylor', 'Moore', 'Jackson', 'Martin', 'Lee', 'Perez', 'Thompson',
    'White', 'Harris', 'Sanchez', 'Clark', 'Ramirez', 'Lewis', 'Robinson', 'Walker',
    'Young', 'Allen', 'King', 'Wright', 'Scott', 'Torres', 'Nguyen', 'Hill', 'Flores',
    'Green', 'Adams', 'Nelson', 'Baker', 'Hall', 'Rivera', 'Campbell', 'Mitchell',
    'Carter', 'Roberts', 'Turner', 'Phillips', 'Evans', 'Collins', 'Stewart',
    'Morris', 'Rogers', 'Reed', 'Cook', 'Morgan', 'Bell', 'Murphy', 'Bailey',
    'Cooper', 'Richardson', 'Cox', 'Howard', 'Ward', 'Peterson', 'Gray',
    'James', 'Watson', 'Brooks', 'Kelly', 'Sanders', 'Price', 'Bennett',
    'Wood', 'Barnes', 'Ross', 'Henderson', 'Coleman', 'Jenkins', 'Perry', 'Powell',
    'Long', 'Patterson', 'Hughes', 'Washington', 'Butler', 'Simmons',
    'Foster', 'Gonzales', 'Bryant', 'Alexander', 'Russell', 'Griffin', 'Diaz',
    'Hayes', 'Myers', 'Ford', 'Hamilton', 'Graham', 'Sullivan', 'Wallace', 'Woods',
    'Cole', 'West', 'Jordan', 'Owens', 'Reynolds', 'Fisher', 'Ellis', 'Harrison',
    'Gibson', 'Mcdonald', 'Cruz', 'Marshall', 'Ortiz', 'Gomez', 'Murray', 'Freeman',
    'Wells', 'Webb', 'Simpson', 'Stevens', 'Tucker', 'Porter', 'Hunter', 'Hicks',
    'Crawford', 'Henry', 'Boyd', 'Mason', 'Morales', 'Kennedy', 'Warren', 'Dixon',
    'Ramos', 'Reyes', 'Burns', 'Gordon', 'Shaw', 'Holmes', 'Rice', 'Robertson',
    'Hunt', 'Black', 'Daniels', 'Palmer', 'Mills', 'Nichols', 'Grant', 'Knight',
    'Ferguson', 'Rose', 'Stone', 'Hawkins', 'Dunn', 'Perkins', 'Hudson', 'Spencer',
    'Gardner', 'Stephens', 'Payne', 'Pierce', 'Berry', 'Matthews', 'Arnold', 'Wagner',
    'Willis', 'Ray', 'Watkins', 'Moreno', 'Snyder', 'Hoover', 'Townsend', 'Medina',
    # additional diverse surnames
    'Patel', 'Shah', 'Kumar', 'Singh', 'Sharma', 'Gupta', 'Mehta', 'Chopra',
    'Nakamura', 'Yamamoto', 'Tanaka', 'Watanabe', 'Suzuki', 'Ito', 'Kato', 'Kobayashi',
    'Kim', 'Park', 'Choi', 'Yoon', 'Lim', 'Jung', 'Kwon', 'Shin',
    'Chen', 'Wang', 'Zhang', 'Liu', 'Yang', 'Huang', 'Zhou', 'Wu',
    'Okafor', 'Adeyemi', 'Nwosu', 'Obi', 'Eze', 'Chukwu', 'Nwachukwu', 'Okonkwo',
    'Mensah', 'Asante', 'Boateng', 'Owusu', 'Amponsah', 'Amoah', 'Darko',
    'Johansson', 'Andersen', 'Eriksson', 'Lindqvist', 'Bergstrom', 'Holmberg', 'Strand',
    'Müller', 'Fischer', 'Weber', 'Becker', 'Hoffmann', 'Schäfer', 'Koch', 'Richter',
    'Dubois', 'Lefevre', 'Moreau', 'Laurent', 'Simon', 'Michel', 'Leblanc', 'Fontaine',
    'Rossi', 'Ferrari', 'Esposito', 'Romano', 'Colombo', 'Ricci', 'Marino', 'Greco',
    'Santos', 'Ferreira', 'Oliveira', 'Costa', 'Rodrigues', 'Almeida', 'Nascimento',
    'Ivanov', 'Smirnov', 'Volkov', 'Popov', 'Sokolov', 'Lebedev', 'Kozlov', 'Novikov',
    'O\'Brien', 'O\'Connor', 'Murphy', 'Walsh', 'Byrne', 'Ryan', 'Doyle', 'Fitzgerald',
    'MacLeod', 'MacKenzie', 'MacDonald', 'Fraser', 'Cameron', 'Stewart', 'Robertson',
    'Beaumont', 'Whitfield', 'Thornton', 'Ashworth', 'Blackwood', 'Hartley', 'Pemberton',
    'Vandenberg', 'Hofstadter', 'Kowalski', 'Wojciechowski', 'Wiśniewski', 'Nowak',
    'Nakagawa', 'Hashimoto', 'Ogawa', 'Matsumoto', 'Inoue', 'Kimura', 'Hayashi',
    'Delgado', 'Vargas', 'Castillo', 'Jimenez', 'Romero', 'Alvarez', 'Ruiz', 'Herrera',
]

TITLES = [
    'Chief Executive Officer', 'Chief Financial Officer', 'Chief Operating Officer',
    'Chief Technology Officer', 'Chief Marketing Officer', 'Chief People Officer',
    'Chief Revenue Officer', 'Chief Legal Officer', 'Chief Compliance Officer',
    'Chief Strategy Officer', 'Chief Data Officer', 'Chief Risk Officer',
    'Chief Communications Officer', 'Chief Diversity Officer', 'Chief of Staff',
    'Vice President of Operations', 'Vice President of Sales', 'Vice President of Marketing',
    'Vice President of Business Development', 'Vice President of Finance',
    'Vice President of Human Resources', 'Vice President of Strategy',
    'Vice President of Corporate Affairs', 'Vice President of Stakeholder Engagement',
    'Vice President of Client Services', 'Vice President of Research',
    'Senior Director of Strategy', 'Senior Director of Corporate Affairs',
    'Senior Director of Stakeholder Engagement', 'Senior Director of Communications',
    'Senior Director of Finance', 'Senior Director of Human Resources',
    'Senior Director of Operations', 'Senior Director of Business Development',
    'Senior Director of Research & Analytics', 'Senior Director of Compliance',
    'Director of Public Relations', 'Director of Human Resources',
    'Director of Business Operations', 'Director of Client Success',
    'Director of Enterprise Partnerships', 'Director of Compliance',
    'Director of Corporate Development', 'Director of Financial Planning & Analysis',
    'Director of Talent Management', 'Director of Total Rewards',
    'Director of Compensation & Benefits', 'Director of Organizational Development',
    'Director of Government Relations', 'Director of Investor Relations',
    'Senior Manager of Corporate Initiatives', 'Senior Manager of Advocacy',
    'Senior Manager of Stakeholder Relations', 'Senior Manager of Operations',
    'Senior Manager of Compensation', 'Senior Manager of Benefits',
    'Senior Manager of Talent Acquisition', 'Senior Manager of Learning & Development',
    'Manager of Corporate Communications', 'Manager of Talent Acquisition',
    'Manager of Strategic Partnerships', 'Manager of Client Relations',
    'Manager of Compensation Programs', 'Manager of Benefits Administration',
    'Manager of Workforce Analytics', 'Manager of Corporate Governance',
    'Senior Analyst, Corporate Strategy', 'Senior Analyst, Market Intelligence',
    'Senior Analyst, Financial Planning', 'Senior Analyst, Compensation',
    'Senior Analyst, Workforce Planning', 'Senior Analyst, Organizational Effectiveness',
    'Business Development Associate', 'Corporate Affairs Specialist',
    'Communications Specialist', 'Operations Coordinator', 'Administrative Coordinator',
    'Executive Assistant to the CEO', 'Executive Assistant to the CFO',
    'Associate, Corporate Development', 'Associate, Public Affairs',
    'Associate, Compensation & Benefits', 'Associate, Talent Acquisition',
    'Research Associate', 'Policy Analyst', 'Governance Specialist',
    'Compensation Analyst', 'Benefits Analyst', 'HR Business Partner',
]

DEPARTMENTS = [
    'Executive Leadership', 'Finance & Accounting', 'Human Resources',
    'Corporate Strategy', 'Operations', 'Business Development',
    'Marketing & Communications', 'Legal & Compliance', 'Technology',
    'Client Services', 'Public Affairs', 'Administrative Services',
    'Stakeholder Relations', 'Corporate Development', 'Risk Management',
    'Research & Analytics', 'Investor Relations', 'Government Relations',
    'Total Rewards', 'Talent Management', 'Learning & Development',
    'Organizational Effectiveness', 'Workforce Planning', 'Corporate Governance',
    'Internal Audit', 'Data & Analytics', 'Knowledge Management',
    'Program Management', 'Policy & Advocacy', 'External Affairs',
]


def generate_employee_batch(n=12):
    """Generate a batch of fake employee dicts."""
    employees = []
    # Use a time-based seed for uniqueness across visits, but pick reproducibly within the batch
    rng = random.Random()
    used_emails = set()

    for _ in range(n):
        first = rng.choice(FIRST_NAMES)
        last = rng.choice(LAST_NAMES)

        # Handle email collisions by appending a number
        base_email = f"{first.lower()}.{last.lower()}@acpwb.com"
        email = base_email
        counter = 2
        while email in used_emails:
            email = f"{first.lower()}.{last.lower()}{counter}@acpwb.com"
            counter += 1
        used_emails.add(email)

        seed = hashlib.md5(f"{first}{last}{rng.random()}".encode()).hexdigest()[:16]

        employees.append({
            'first_name': first,
            'last_name': last,
            'email': email,
            'title': rng.choice(TITLES),
            'department': rng.choice(DEPARTMENTS),
            'avatar_seed': seed,
        })

    return sorted(employees, key=lambda e: e['last_name'])
