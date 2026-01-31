package data

import "strings"

// District represents an Indonesian district for autocomplete
type District struct {
	Name     string
	City     string
	Province string
}

// IndonesianCities contains major Indonesian cities for autocomplete
// This enables fuzzy search that Biteship API doesn't support
var IndonesianCities = []string{
	// Major cities - these will be used for city-level autocomplete
	"Jakarta", "Jakarta Pusat", "Jakarta Selatan", "Jakarta Barat", "Jakarta Timur", "Jakarta Utara",
	"Surabaya", "Bandung", "Medan", "Semarang", "Makassar", "Palembang", "Tangerang", "Depok",
	"Bekasi", "Bogor", "Malang", "Yogyakarta", "Solo", "Surakarta", "Denpasar", "Balikpapan",
	"Samarinda", "Pontianak", "Banjarmasin", "Manado", "Padang", "Pekanbaru", "Batam",
	"Cirebon", "Tasikmalaya", "Sukabumi", "Karawang", "Purwokerto", "Tegal", "Pekalongan",
	"Magelang", "Kudus", "Jepara", "Demak", "Kendal", "Ungaran", "Salatiga", "Klaten",
	"Boyolali", "Wonogiri", "Karanganyar", "Sragen", "Blora", "Rembang", "Pati", "Grobogan",
	"Kediri", "Mojokerto", "Pasuruan", "Probolinggo", "Lumajang", "Jember", "Banyuwangi",
	"Situbondo", "Bondowoso", "Madiun", "Ngawi", "Magetan", "Ponorogo", "Pacitan", "Trenggalek",
	"Tulungagung", "Blitar", "Nganjuk", "Lamongan", "Gresik", "Tuban", "Bojonegoro",
	"Serang", "Cilegon", "Pandeglang", "Lebak", "Tangerang Selatan",
	"Cimahi", "Garut", "Subang", "Purwakarta", "Cianjur", "Majalengka", "Kuningan", "Indramayu",
	"Banjar", "Pangandaran", "Sumedang",
	"Lampung", "Bandar Lampung", "Metro", "Palangkaraya", "Jambi", "Bengkulu", "Kendari",
	"Palu", "Gorontalo", "Ambon", "Ternate", "Jayapura", "Sorong", "Manokwari", "Merauke",
	"Kupang", "Mataram", "Bima", "Sumbawa", "Lombok",
	"Aceh", "Banda Aceh", "Lhokseumawe", "Langsa", "Sabang",
}

// IndonesianDistricts contains common district names for autocomplete
// This enables fuzzy search that Biteship API doesn't support
var IndonesianDistricts = []District{
	// SEMARANG (16 kecamatan)
	{Name: "Banyumanik", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Candisari", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Gajahmungkur", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Gayamsari", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Genuk", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Gunungpati", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Mijen", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Ngaliyan", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Pedurungan", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Semarang Barat", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Semarang Selatan", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Semarang Tengah", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Semarang Timur", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Semarang Utara", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Tembalang", City: "Semarang", Province: "Jawa Tengah"},
	{Name: "Tugu", City: "Semarang", Province: "Jawa Tengah"},
}


// More districts - JAKARTA
var JakartaDistricts = []District{
	// Jakarta Pusat (8 kecamatan)
	{Name: "Cempaka Putih", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Gambir", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Johar Baru", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Kemayoran", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Menteng", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Sawah Besar", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Senen", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	{Name: "Tanah Abang", City: "Jakarta Pusat", Province: "DKI Jakarta"},
	// Jakarta Selatan (10 kecamatan)
	{Name: "Cilandak", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Jagakarsa", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Kebayoran Baru", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Kebayoran Lama", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Mampang Prapatan", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Pancoran", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Pasar Minggu", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Pesanggrahan", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Setiabudi", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	{Name: "Tebet", City: "Jakarta Selatan", Province: "DKI Jakarta"},
	// Jakarta Barat (8 kecamatan)
	{Name: "Cengkareng", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Grogol Petamburan", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Kalideres", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Kebon Jeruk", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Kembangan", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Palmerah", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Taman Sari", City: "Jakarta Barat", Province: "DKI Jakarta"},
	{Name: "Tambora", City: "Jakarta Barat", Province: "DKI Jakarta"},
	// Jakarta Timur (10 kecamatan)
	{Name: "Cakung", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Cipayung", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Ciracas", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Duren Sawit", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Jatinegara", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Kramat Jati", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Makasar", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Matraman", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Pasar Rebo", City: "Jakarta Timur", Province: "DKI Jakarta"},
	{Name: "Pulo Gadung", City: "Jakarta Timur", Province: "DKI Jakarta"},
	// Jakarta Utara (6 kecamatan)
	{Name: "Cilincing", City: "Jakarta Utara", Province: "DKI Jakarta"},
	{Name: "Kelapa Gading", City: "Jakarta Utara", Province: "DKI Jakarta"},
	{Name: "Koja", City: "Jakarta Utara", Province: "DKI Jakarta"},
	{Name: "Pademangan", City: "Jakarta Utara", Province: "DKI Jakarta"},
	{Name: "Penjaringan", City: "Jakarta Utara", Province: "DKI Jakarta"},
	{Name: "Tanjung Priok", City: "Jakarta Utara", Province: "DKI Jakarta"},
}


// BANDUNG districts
var BandungDistricts = []District{
	{Name: "Andir", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Antapani", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Arcamanik", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Astana Anyar", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Babakan Ciparay", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Bandung Kidul", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Bandung Kulon", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Bandung Wetan", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Batununggal", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Bojongloa Kaler", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Bojongloa Kidul", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Buahbatu", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Cibeunying Kaler", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Cibeunying Kidul", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Cibiru", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Cicendo", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Cidadap", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Cinambo", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Coblong", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Gedebage", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Kiaracondong", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Lengkong", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Mandalajati", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Panyileukan", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Rancasari", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Regol", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Sukajadi", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Sukasari", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Sumur Bandung", City: "Bandung", Province: "Jawa Barat"},
	{Name: "Ujung Berung", City: "Bandung", Province: "Jawa Barat"},
}

// SURABAYA districts
var SurabayaDistricts = []District{
	{Name: "Asemrowo", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Benowo", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Bubutan", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Bulak", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Dukuh Pakis", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Gayungan", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Genteng", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Gubeng", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Gunung Anyar", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Jambangan", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Karang Pilang", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Kenjeran", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Krembangan", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Lakarsantri", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Mulyorejo", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Pabean Cantikan", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Pakal", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Rungkut", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Sambikerep", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Sawahan", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Semampir", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Simokerto", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Sukolilo", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Sukomanunggal", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Tambaksari", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Tandes", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Tegalsari", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Tenggilis Mejoyo", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Wiyung", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Wonocolo", City: "Surabaya", Province: "Jawa Timur"},
	{Name: "Wonokromo", City: "Surabaya", Province: "Jawa Timur"},
}


// GetAllDistricts returns all districts combined
func GetAllDistricts() []District {
	all := make([]District, 0, 200)
	all = append(all, IndonesianDistricts...)
	all = append(all, JakartaDistricts...)
	all = append(all, BandungDistricts...)
	all = append(all, SurabayaDistricts...)
	return all
}

// SearchDistricts searches districts by query (fuzzy match)
func SearchDistricts(query string) []District {
	query = strings.ToLower(strings.TrimSpace(query))
	if len(query) < 2 {
		return nil
	}

	var results []District
	all := GetAllDistricts()

	for _, d := range all {
		nameLower := strings.ToLower(d.Name)
		cityLower := strings.ToLower(d.City)
		
		// Match if query is prefix of name or city, or contained in name
		if strings.HasPrefix(nameLower, query) ||
			strings.HasPrefix(cityLower, query) ||
			strings.Contains(nameLower, query) {
			results = append(results, d)
		}
	}

	return results
}

// SearchCities searches cities by query (fuzzy match)
// Returns matching city names for autocomplete
func SearchCities(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if len(query) < 2 {
		return nil
	}

	var results []string
	seen := make(map[string]bool)

	for _, city := range IndonesianCities {
		cityLower := strings.ToLower(city)
		
		// Match if query is prefix or contained in city name
		if strings.HasPrefix(cityLower, query) || strings.Contains(cityLower, query) {
			if !seen[city] {
				seen[city] = true
				results = append(results, city)
			}
		}
	}

	return results
}
