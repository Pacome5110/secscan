# SecScan Deployment Guide (F10)

Bu doküman, SecScan projesini production ortamına (örn: Vercel + Railway) nasıl deploy edeceğinizi açıklar.

## 1. Backend Deploy (Railway / Render.com / Fly.io)

Go backend uygulamasını deploy etmek için projede bulunan çok aşamalı (multi-stage) `backend/Dockerfile` kullanılmalıdır.

1. Railway (veya Render) üzerinde yeni bir "New Project -> Deploy from Git repo" seçeneğini tıklayın.
2. Repo dizini olarak `backend/` veya `.` seçin.
3. Railway otomatik olarak Dockerfile'ı tanıyacak ve Go uygulamasını derleyip ayağa kaldıracaktır.
4. Çıkan public API adresini (örn: `https://secscan-api.up.railway.app`) not alın.

---

## 2. Frontend Deploy (Vercel / Netlify)

Next.js frontend uygulamasını Vercel veya benzeri Next.js destekli bir servise deploy edeceğiz.

1. Vercel dashboard'ına girin ve "Add New -> Project" deyin.
2. SecScan deponuzu seçin.
3. **Framework Preset** olarak otomatik "Next.js" seçilecektir.
4. **Root Directory** olarak `frontend` klasörünü seçin.
5. **Environment Variables** (Çevresel Değişkenler) alanına aşağıdaki değişkeni ekleyin:
   ```env
   NEXT_PUBLIC_API_URL=https://secscan-api.up.railway.app
   ```
   *(Buraya adım 1'den aldığınız backend URL'sini koymalısınız. Kod içindeki göreli yol (`/api/scan/...`) üretim ortamında çalışmazsa, Next.js rewrite kuralı veya doğrudan bu değişken kullanılabilir).*

6. "Deploy" butonuna basın.

---

## 3. Demo ve Test İzlemi

- Frontend adresinize gidin (örn: `https://secscan.vercel.app`).
- Hedef URL girin (Örn: `http://localhost:3000` değil, public bir test sitesi: `http://testphp.vulnweb.com`).
- Taramayı başlatın. Frontend, backend'inize bağlanacak, Go rutinleri paralel olarak sitenin portlarını, TLS'ini, fuzzer ve güvenlik başlıklarını test edecektir.
- En son ekranda "Güvenlik Skoru Dağılımı" radar şemasını görüntüleyin ve `PDF İndir` butonu ile raporu test edin.

**İyi taramalar! 🛡️**
