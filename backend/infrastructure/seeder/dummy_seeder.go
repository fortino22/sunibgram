package seeder

import (
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	facultymodel "github.com/sunibgram/domain/model/faculty"
	rolemodel "github.com/sunibgram/domain/model/role"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
)

type dummyAccount struct {
	Name         string
	Username     string
	Email        string
	Password     string
	RoleName     string
	Gender       userprofilemodel.Gender
	BirthdayDate time.Time
	FacultyName  string
}

var dummyAccounts = []dummyAccount{
	{Name: "Octa", Username: "Octa.hi", Email: "octa@binus.ac.id", Password: "dummy123", RoleName: "admin", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2003, 4, 12, 0, 0, 0, 0, time.UTC), FacultyName: "School of Computer Science"},
	{Name: "Devin", Username: "devin.hi", Email: "devin@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderMale, BirthdayDate: time.Date(2002, 8, 25, 0, 0, 0, 0, time.UTC), FacultyName: "School of Information Systems"},
	{Name: "chermiese", Username: "chermiese", Email: "chermiese@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2003, 2, 19, 0, 0, 0, 0, time.UTC), FacultyName: "School of Design"},
	{Name: "Aldi", Username: "aldi.hi", Email: "aldi@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderMale, BirthdayDate: time.Date(2002, 1, 14, 0, 0, 0, 0, time.UTC), FacultyName: "School of Computer Science"},
	{Name: "Bella", Username: "bella.hi", Email: "bella@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2003, 6, 3, 0, 0, 0, 0, time.UTC), FacultyName: "School of Information Systems"},
	{Name: "Caesar", Username: "caesar.hi", Email: "caesar@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderMale, BirthdayDate: time.Date(2001, 9, 21, 0, 0, 0, 0, time.UTC), FacultyName: "School of Accounting"},
	{Name: "Dina", Username: "dina.hi", Email: "dina@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2003, 12, 30, 0, 0, 0, 0, time.UTC), FacultyName: "School of Design"},
	{Name: "Eko", Username: "eko.hi", Email: "eko@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderMale, BirthdayDate: time.Date(2002, 3, 8, 0, 0, 0, 0, time.UTC), FacultyName: "Faculty of Engineering"},
	{Name: "Fira", Username: "fira.hi", Email: "fira@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2003, 7, 17, 0, 0, 0, 0, time.UTC), FacultyName: "Faculty of Humanities"},
	{Name: "Gilang", Username: "gilang.hi", Email: "gilang@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderMale, BirthdayDate: time.Date(2001, 11, 2, 0, 0, 0, 0, time.UTC), FacultyName: "School of Computing and Creative Arts"},
	{Name: "Hana", Username: "hana.hi", Email: "hana@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2002, 5, 26, 0, 0, 0, 0, time.UTC), FacultyName: "School of Information Systems"},
	{Name: "Irfan", Username: "irfan.hi", Email: "irfan@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderMale, BirthdayDate: time.Date(2003, 10, 11, 0, 0, 0, 0, time.UTC), FacultyName: "School of Computer Science"},
	{Name: "Jihan", Username: "jihan.hi", Email: "jihan@binus.ac.id", Password: "dummy123", RoleName: "user", Gender: userprofilemodel.GenderFemale, BirthdayDate: time.Date(2002, 4, 5, 0, 0, 0, 0, time.UTC), FacultyName: "BINUS Business School"},
}

func SeedDummy(db *gorm.DB) {
	roleCache := make(map[string]rolemodel.Role)
	facultyCache := make(map[string]facultymodel.Faculty)

	for _, acc := range dummyAccounts {
		acc.Username = strings.TrimSpace(acc.Username)
		acc.Email = strings.TrimSpace(acc.Email)

		var existing usermodel.User
		if err := db.Where("email = ?", acc.Email).First(&existing).Error; err == nil {
			var existingProfile userprofilemodel.UserProfile
			if err := db.Where("user_id = ?", existing.UserID).First(&existingProfile).Error; err != nil {
				db.Create(&userprofilemodel.UserProfile{
					UserID:       existing.UserID,
					Gender:       acc.Gender,
					BirthdayDate: acc.BirthdayDate,
				})
			}
			var existingSettings usersettingsmodel.UserSettings
			if err := db.Where("user_id = ?", existing.UserID).First(&existingSettings).Error; err != nil {
				db.Create(&usersettingsmodel.UserSettings{UserID: existing.UserID})
			}
			var existingFaculty usermodel.UserFaculty
			if err := db.Where("user_id = ?", existing.UserID).First(&existingFaculty).Error; err != nil {
				faculty, err := resolveFaculty(db, facultyCache, acc.FacultyName, acc.Email)
				if err == nil {
					db.Create(&usermodel.UserFaculty{UserID: existing.UserID, FacultyID: faculty.FacultyID})
				}
			}
			continue
		}

		role, ok := roleCache[acc.RoleName]
		if !ok {
			if err := db.Where("name = ?", acc.RoleName).First(&role).Error; err != nil {
				log.Printf("dummy seeder: role '%s' not found, skipping %s", acc.RoleName, acc.Email)
				continue
			}
			roleCache[acc.RoleName] = role
		}

		faculty, err := resolveFaculty(db, facultyCache, acc.FacultyName, acc.Email)
		if err != nil {
			continue
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("dummy seeder: failed to hash password for %s: %v", acc.Email, err)
			continue
		}

		user := usermodel.User{
			Name:     acc.Name,
			UserName: acc.Username,
			Email:    acc.Email,
			Password: string(hashed),
			RoleID:   role.RoleID,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("dummy seeder: failed to create user %s: %v", acc.Email, err)
			continue
		}

		if err := db.Create(&userprofilemodel.UserProfile{
			UserID:       user.UserID,
			Gender:       acc.Gender,
			BirthdayDate: acc.BirthdayDate,
		}).Error; err != nil {
			log.Printf("dummy seeder: failed to create profile for %s: %v", acc.Email, err)
			continue
		}

		if err := db.Create(&usersettingsmodel.UserSettings{UserID: user.UserID}).Error; err != nil {
			log.Printf("dummy seeder: failed to create settings for %s: %v", acc.Email, err)
			continue
		}

		if err := db.Create(&usermodel.UserFaculty{UserID: user.UserID, FacultyID: faculty.FacultyID}).Error; err != nil {
			log.Printf("dummy seeder: failed to create faculty for %s: %v", acc.Email, err)
			continue
		}

		log.Printf("dummy seeder: created user %s (%s)", acc.Username, acc.Email)
	}
}

func resolveFaculty(db *gorm.DB, cache map[string]facultymodel.Faculty, name, email string) (facultymodel.Faculty, error) {
	if f, ok := cache[name]; ok {
		return f, nil
	}
	var faculty facultymodel.Faculty
	if err := db.Where("faculty_name = ?", name).First(&faculty).Error; err != nil {
		log.Printf("dummy seeder: faculty '%s' not found, skipping %s", name, email)
		return faculty, err
	}
	cache[name] = faculty
	return faculty, nil
}
