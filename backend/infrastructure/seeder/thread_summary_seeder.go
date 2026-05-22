package seeder

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	threadmodel "github.com/sunibgram/domain/model/thread"
	usermodel "github.com/sunibgram/domain/model/user"
)

var sampleReplies = []string{
	"This is really helpful, thanks for sharing!",
	"I agree with this completely, been feeling the same way.",
	"Can someone explain this more? I'm a bit confused.",
	"This happened to me last semester too, it's so frustrating.",
	"Anyone know if this affects the final exam schedule?",
	"I already asked the lecturer and they said it's fine.",
	"Not sure I agree with this take honestly.",
	"Same experience here at the Alam Sutera campus.",
	"The library has resources on this if anyone needs them.",
	"Has anyone tried reaching out to student affairs about this?",
	"This is exactly why I joined this platform, great discussion.",
	"Binus should really address this more officially.",
	"I heard they're planning to change this next semester.",
	"Does this apply to online students too or just campus?",
	"My friend in another faculty said the same thing.",
	"Just share the assignment link here if you want feedback.",
	"I think the bigger issue is the grading rubric itself.",
	"Wait this is news to me, when did this happen?",
	"The announcement was posted in the student portal last week.",
	"Why does nobody check the portal, all info is there lol.",
	"This reminds me of what happened with the attendance system.",
	"Honestly the system needs a complete overhaul.",
	"Agreed, but let's stay constructive about it.",
	"Anyone forming a study group for this? Drop your LINE.",
	"I already raised this in class and the lecturer ignored it.",
	"Maybe we should escalate this to the dean's office.",
	"That seems extreme, let's try talking to the coordinator first.",
	"The coordinator is never available during office hours anyway.",
	"Book an appointment via the academic portal, it works.",
	"I did that and waited two weeks for a response.",
	"Classic Binus admin speed right there.",
	"In all seriousness though this is a real problem.",
	"Does anyone have the official policy document for this?",
	"Check the academic handbook on the student portal.",
	"The handbook hasn't been updated since 2022 though.",
	"That's a whole other issue we should discuss.",
	"Focus guys, the original point is still valid.",
	"Sorry you're right, back to the main topic.",
	"I think the solution is simpler than everyone is making it.",
	"Please elaborate, I want to understand your point.",
	"Basically just follow the SOP and document everything.",
	"That's easy to say but the SOP is unclear on this.",
	"Agreed, the SOP has gaps that need to be addressed.",
	"Semester is almost over anyway, hopefully next sem is better.",
	"This thread really blew up, glad people are talking about it.",
	"Upvoting this so more people see it.",
	"Sharing this with my groupmates, very relevant.",
	"Thanks OP for bringing this up, needed to be said.",
	"Let's keep the conversation going and follow up next week.",
}

var sampleRepliesID = []string{
	"Gue pernah ngalamin kejadian gak enak sama salah satu senior di kampus, tapi gak berani lapor karena takut dianggap lebay.",
	"Ini harus diangkat. Banyak korban yang diem karena malu atau takut nama baiknya hancur.",
	"Kampus harusnya punya mekanisme pelaporan yang aman dan anonim, bukan cuma kotak saran yang gak jelas.",
	"Temen gue pernah dilecehkan secara verbal waktu ospek, pelakunya seniornya sendiri. Gak ada yang mau denger.",
	"Satgas PPKS di Binus tuh aktif gak sih? Pernah ada yang lapor ke sana dan hasilnya gimana?",
	"Gue udah lapor ke satgas semester lalu. Prosesnya lama dan kurang transparan, tapi akhirnya ada tindakan.",
	"Masalahnya banyak yang gak tau kalau ada satgas PPKS. Sosialisasinya kurang banget.",
	"Setuju. Gue baru tau ada satgas itu setelah baca thread ini. Harusnya diumumin lebih masif.",
	"Pelaku sering orang yang punya 'power' di kampus, makanya korban takut speak up.",
	"Betul banget. Ada kasus dosen yang bersikap gak pantas ke mahasiswi tapi dilindungi institusi.",
	"Ini bukan cuma soal pelecehan fisik ya, cat calling dan komentar seksual juga masuk kategori pelecehan.",
	"Banyak mahasiswa gak sadar kalau yang mereka anggap 'bercanda' itu udah masuk pelecehan.",
	"Edukasi consent dan batas perilaku harus masuk kurikulum wajib atau setidaknya orientasi mahasiswa baru.",
	"Gue pernah jadi saksi tapi gak tau harus ngapain. Ada gak panduan buat saksi yang mau lapor?",
	"Ada, bisa lewat satgas PPKS atau langsung ke BK. Tapi emang kurang disosialisasiin.",
	"Korban juga sering disalahin, dipertanyain pakaiannya atau sikapnya. Victim blaming parah di sini.",
	"Itu yang bikin orang males lapor. Bukannya dibela malah dikambing hitamkan.",
	"Kampus harus jamin kerahasiaan identitas pelapor supaya orang mau speak up tanpa takut.",
	"Penting banget. Gue gak mau lapor karena khawatir nama gue tersebar dan malah kena bully.",
	"Komunitas mahasiswa juga harus aktif bikin ruang aman, jangan cuma nunggu kampus bertindak.",
	"Setuju. Sesama mahasiswa harus saling jaga, bukan nonton doang kalau ada yang gak bener.",
	"Ada yang tau nomor atau kontak resmi satgas PPKS Binus? Minta disini biar bisa disebarluasin.",
	"Coba cek di website Binus bagian kemahasiswaan, harusnya ada kontaknya di sana.",
	"Udah dicek, kontaknya ada tapi emailnya lama bales. Perlu ada hotline yang lebih responsif.",
	"Ini masalah serius yang butuh perhatian lebih dari pihak rektorat, bukan cuma dilempar ke satgas doang.",
	"Rektorat harus buat kebijakan zero tolerance yang jelas dan diumumin ke seluruh civitas akademik.",
	"Kebijakan ada, implementasinya yang sering gak konsisten. Kasus serupa beda treatment-nya.",
	"Bener, ada yang diproses cepet ada yang dipeti-eskan. Tergantung siapa pelakunya.",
	"Transparansi proses penanganan kasus itu penting supaya mahasiswa percaya sistemnya jalan.",
	"Gue harap ada laporan tahunan soal jumlah kasus yang masuk dan penanganannya, kayak kampus luar negeri.",
	"Ide bagus. Akuntabilitas publik bisa jadi pressure buat kampus beneran nangani kasus dengan serius.",
	"Jangan lupa dukungan psikologis buat korban. Traumanya gak hilang cuma karena kasusnya selesai.",
	"Binus punya layanan konseling gratis kan? Semoga korban yang butuh bantuan mau akses layanan itu.",
	"Layanan konselingnya ada tapi antriannya panjang dan jam operasionalnya terbatas banget.",
	"Harus ditambah jumlah konselor dan jam layanannya, ini kebutuhan mendesak.",
	"Support group antar mahasiswa juga bisa jadi alternatif sambil nunggu sistem kampus diperbaiki.",
	"Ada yang mau bantu inisiasi support group? Gue siap ikut kalau ada yang mau gerakin.",
	"Gue juga. Mungkin bisa mulai dari komunitas kecil dulu di tiap fakultas.",
	"Kita butuh lebih banyak edukasi soal ini di media sosial kampus, jangan cuma konten yang fun-fun aja.",
	"Thread ini harusnya di-pin atau dibagiin ke akun resmi Binus biar lebih banyak yang sadar.",
	"Setuju, ini penting banget. Awareness adalah langkah pertama.",
	"Terima kasih udah berani angkat topik ini. Gak mudah tapi perlu.",
	"Semoga korban yang baca thread ini tau mereka gak sendirian dan ada yang peduli.",
	"Kalau ada yang butuh cerita atau minta saran, DM aja. Gue siap dengerin.",
	"Keberanian speak up itu butuh dukungan komunitas. Kita harus jadi komunitas yang aman buat semua.",
	"Jangan berhenti bahas ini cuma di media sosial, bawa ke forum resmi mahasiswa juga.",
	"BEM harusnya jadiin ini agenda prioritas, bukan cuma sibuk acara dan lomba.",
	"Setuju. Advokasi hak mahasiswa itu salah satu fungsi utama BEM.",
	"Gue akan angkat ini di rapat BEM minggu depan. Minta dukungan kalian semua.",
	"Kita dukung penuh. Ubah diskusi jadi aksi nyata, jangan cuma berhenti di kolom komentar.",
}

func seedThread(db *gorm.DB, owner usermodel.User, content, marker string, replies []string) {
	var existing threadmodel.Thread
	if err := db.Where("content LIKE ? AND parent_id IS NULL", "%"+marker+"%").First(&existing).Error; err == nil {
		log.Printf("thread summary seeder: seed thread '%s' already exists (id=%d), skipping", marker, existing.ThreadID)
		return
	}

	root := threadmodel.Thread{
		UserID:       owner.UserID,
		Type:         threadmodel.ThreadTypePost,
		Content:      fmt.Sprintf("%s %s", content, marker),
		CommentCount: len(replies),
	}
	if err := db.Create(&root).Error; err != nil {
		log.Printf("thread summary seeder: failed to create root thread '%s': %v", marker, err)
		return
	}

	for i, text := range replies {
		reply := threadmodel.Thread{
			UserID:   owner.UserID,
			Type:     threadmodel.ThreadTypeReply,
			Content:  text,
			ParentID: &root.ThreadID,
			RootID:   &root.ThreadID,
		}
		if err := db.Create(&reply).Error; err != nil {
			log.Printf("thread summary seeder: failed to create reply %d for '%s': %v", i+1, marker, err)
		}
	}

	log.Printf("thread summary seeder: created thread id=%d ('%s') with %d replies", root.ThreadID, marker, len(replies))
}

func SeedThreadWithReplies(db *gorm.DB) {
	var owner usermodel.User
	if err := db.Where("email = ?", "fortino@binus.ac.id").First(&owner).Error; err != nil {
		log.Printf("thread summary seeder: user fortino@binus.ac.id not found, run SeedDummy first")
		return
	}

	seedThread(db, owner,
		"What do you think about student life at Binus? Share your experience!",
		"[summary-seed-en]",
		sampleReplies,
	)

	seedThread(db, owner,
		"Pelecehan seksual di kampus masih sering terjadi tapi jarang dilapor. Yuk kita bahas dan cari solusinya bareng.",
		"[summary-seed-id]",
		sampleRepliesID,
	)

	if err := db.Exec(`
		UPDATE threads
		SET feed_score = (
			like_count::float +
			3.0 * comment_count::float +
			5.0 * repost_count::float +
			4.0 * bookmark_count::float
		) / (
			1.0 + POWER(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600.0, 1.5)
		)
		WHERE created_at > NOW() - INTERVAL '7 days'
	`).Error; err != nil {
		log.Printf("thread summary seeder: feed score recalc failed: %v", err)
	}
}
