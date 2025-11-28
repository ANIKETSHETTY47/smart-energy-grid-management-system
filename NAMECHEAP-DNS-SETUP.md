# 🌐 Namecheap DNS Setup for dashboard.aniketshetty.me

## Quick Reference

**Your Domain:** aniketshetty.me  
**Subdomain:** dashboard.aniketshetty.me  
**Target:** [Your Elastic Beanstalk URL from deployment]

---

## 📋 Step-by-Step Instructions

### Step 1: Log in to Namecheap

1. Go to: **https://www.namecheap.com/myaccount/login**
2. Enter your credentials
3. Click **Sign In**

---

### Step 2: Navigate to Domain Management

1. Click **Domain List** in the left sidebar
2. Find **aniketshetty.me** in your domain list
3. Click the **MANAGE** button next to it

---

### Step 3: Access Advanced DNS

1. Click the **Advanced DNS** tab at the top
2. You'll see a list of existing DNS records

---

### Step 4: Add CNAME Record

1. Click **ADD NEW RECORD** button

2. Fill in the form:
   ```
   Type: CNAME Record
   Host: dashboard
   Value: [Paste your Elastic Beanstalk URL here]
   TTL: Automatic (or select 5 min / 300)
   ```

3. Click the **✓** (checkmark) or **Save** button

---

## ✅ What It Should Look Like

After adding, you should see:

| Type | Host | Value | TTL |
|------|------|-------|-----|
| CNAME Record | dashboard | energy-dashboard-env.eba-xxxxx.eu-north-1.elasticbeanstalk.com | Automatic |

---

## ⏰ DNS Propagation Time

**Normal:** 5-30 minutes  
**Maximum:** Up to 60 minutes  

**Why does it take time?**
- DNS changes need to propagate across the internet
- Different DNS servers update at different times
- Namecheap usually propagates quickly (10-20 minutes)

---

## 🧪 How to Test DNS Propagation

### Method 1: Using Terminal (Mac)

```bash
# Check if DNS is working
nslookup dashboard.aniketshetty.me

# If working, you'll see:
# Non-authoritative answer:
# dashboard.aniketshetty.me  canonical name = energy-dashboard-env.eba-xxxxx.elasticbeanstalk.com
```

### Method 2: Using Online Tool

1. Go to: **https://www.whatsmydns.net/**
2. Enter: `dashboard.aniketshetty.me`
3. Select: `CNAME` from dropdown
4. Click **Search**
5. Green checkmarks = DNS propagated in that location

### Method 3: Try Opening in Browser

Simply try opening:
```
http://dashboard.aniketshetty.me
```

If it loads your dashboard = DNS is working!

---

## 🎯 Complete Example

**Before DNS Setup:**
```
http://energy-dashboard-env.eba-xxxxx.eu-north-1.elasticbeanstalk.com
                    ↓
               (Works but long URL)
```

**After DNS Setup:**
```
http://dashboard.aniketshetty.me
                    ↓
         (Short, professional URL!)
```

---

## 🆘 Troubleshooting

### Issue 1: "Record already exists"

**Solution:**
- Look for existing `dashboard` record
- Click **Edit** (pencil icon)
- Update the Value to your new Elastic Beanstalk URL
- Save

---

### Issue 2: DNS not propagating after 1 hour

**Solution:**
1. Check the CNAME record is correct (no typos)
2. Verify the Host is exactly: `dashboard` (not `dashboard.aniketshetty.me`)
3. Verify the Value has the full EB URL (with `.elasticbeanstalk.com`)
4. Try flushing your local DNS cache:
   ```bash
   sudo dscacheutil -flushcache
   sudo killall -HUP mDNSResponder
   ```

---

### Issue 3: "This site can't be reached"

**Possible causes:**

1. **DNS not propagated yet**
   - Wait longer (up to 60 minutes)
   - Check with whatsmydns.net

2. **Elastic Beanstalk environment not ready**
   - Go to AWS Console
   - Check if environment health is GREEN
   - Verify the URL works directly

3. **Wrong CNAME value**
   - Value should NOT include `http://`
   - Should be just: `energy-dashboard-env.eba-xxxxx.elasticbeanstalk.com`

---

## 📸 Screenshots for Your Report

Take these screenshots:

### 1. Namecheap DNS Settings
- Show the CNAME record you created
- Shows professional domain configuration

### 2. DNS Propagation Check
- Screenshot from whatsmydns.net showing green checkmarks
- Proves DNS is working globally

### 3. Working Dashboard
- Browser showing `http://dashboard.aniketshetty.me`
- All dashboard pages accessible

---

## 🎓 For Your Project Report

### Sample Text:

> "The frontend dashboard was deployed to AWS Elastic Beanstalk and configured with a custom domain using CNAME DNS records. The application is publicly accessible at http://dashboard.aniketshetty.me without requiring any authentication or AWS account access.
>
> DNS configuration was implemented through Namecheap's Advanced DNS settings, creating a CNAME record that points the subdomain 'dashboard.aniketshetty.me' to the Elastic Beanstalk environment URL. This provides a professional, memorable URL for accessing the application while maintaining the scalability and reliability of AWS infrastructure."

---

## ⚡ Quick Command Reference

```bash
# Check DNS status
nslookup dashboard.aniketshetty.me

# Test if site is reachable
curl -I http://dashboard.aniketshetty.me

# Flush local DNS cache (if needed)
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder

# Check Elastic Beanstalk environment status
aws elasticbeanstalk describe-environments \
  --application-name energy-dashboard \
  --environment-names energy-dashboard-env \
  --region eu-north-1
```

---

## ✅ Success Checklist

Before considering setup complete:

- [ ] Logged into Namecheap
- [ ] Found aniketshetty.me domain
- [ ] Opened Advanced DNS tab
- [ ] Added CNAME record (dashboard → EB URL)
- [ ] Saved the record
- [ ] Waited 10-60 minutes
- [ ] Tested with nslookup (shows CNAME)
- [ ] Opened http://dashboard.aniketshetty.me in browser
- [ ] Dashboard loads successfully
- [ ] All pages work (dashboard, alerts, equipment, analytics)
- [ ] Took screenshots for report

---

## 🎉 When Everything Works

You should be able to:

1. ✅ Open `http://dashboard.aniketshetty.me` in any browser
2. ✅ See your Energy Grid Dashboard homepage
3. ✅ Navigate to all pages without issues
4. ✅ Dashboard connects to backend API successfully
5. ✅ Anyone can access it (no login required)

**Perfect for your project demo! 🚀**

---

## 💡 Pro Tips

1. **Test both URLs:**
   - Direct EB URL (should work immediately)
   - Custom domain (works after DNS propagates)

2. **Use incognito/private browsing:**
   - Avoids cached DNS issues
   - Shows what examiners will see

3. **Mobile testing:**
   - Test on phone to verify DNS propagation
   - Shows professional, responsive design

4. **Keep EB URL as backup:**
   - Include both URLs in your report
   - If DNS issues arise, you have fallback

---

## 📞 Need Help?

If stuck:

1. **Check AWS Console:**
   - Is EB environment GREEN?
   - Does direct EB URL work?

2. **Check DNS:**
   - Use whatsmydns.net
   - Use nslookup command

3. **Check logs:**
   ```bash
   aws logs tail /aws/elasticbeanstalk/energy-dashboard-env/var/log/web.stdout.log \
     --follow --region eu-north-1
   ```

4. **Verify Namecheap settings:**
   - Re-check the CNAME record
   - Make sure no typos

---

*Last Updated: November 23, 2024*  
*Domain: aniketshetty.me via Namecheap*  
*Target: AWS Elastic Beanstalk (eu-north-1)*
