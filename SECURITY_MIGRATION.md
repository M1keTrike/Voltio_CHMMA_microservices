# Security Migration Guide

## Important Security Notice

**Action Required:** If you were using a previous version of this codebase with hardcoded credentials, please follow these steps immediately.

## What Changed?

Previous versions of this codebase contained hardcoded credentials including:
- RabbitMQ connection URIs with usernames and passwords
- InfluxDB authentication tokens  
- PostgreSQL database credentials
- API webhook tokens

**These credentials have been removed** and replaced with environment variable configuration.

## Immediate Actions Required

### 1. Rotate All Credentials

If you were using the hardcoded credentials from previous versions, you should:

1. **Change RabbitMQ credentials:**
   - Log into your RabbitMQ management interface
   - Change the password for the affected user accounts
   - Update your `.env` file with the new credentials

2. **Rotate InfluxDB tokens:**
   - Generate new authentication tokens in InfluxDB
   - Revoke old tokens
   - Update your `.env` file with new tokens

3. **Update PostgreSQL credentials:**
   - Change database passwords
   - Update connection strings in `.env` files

4. **Regenerate API tokens:**
   - Create new webhook tokens
   - Update API configurations
   - Update `.env` files

### 2. Set Up Environment Variables

1. Copy the example files:
   ```bash
   cp .env.example .env
   cp Backend/automation-engine/.env.example Backend/automation-engine/.env
   ```

2. Fill in your **NEW** credentials (after rotating them per step 1)

3. Verify `.env` files are listed in `.gitignore`

### 3. Verify Security

1. Check that no `.env` files are tracked by git:
   ```bash
   git status
   ```

2. Ensure no credentials remain in git history:
   ```bash
   git log --all --full-history -- "*/.env"
   ```

3. If you find `.env` files in history, consider:
   - Using `git filter-branch` or `BFG Repo-Cleaner` to remove them
   - Contacting your security team
   - Following your organization's incident response procedures

### 4. Update Deployment Configurations

If you have CI/CD pipelines or deployment scripts:

1. Update them to use environment variables
2. Use secure secret management (e.g., GitHub Secrets, AWS Secrets Manager, HashiCorp Vault)
3. Never log or display credential values
4. Ensure environment variables are properly scoped

## Verification Checklist

- [ ] All credentials have been rotated
- [ ] New credentials are stored in `.env` files (not in code)
- [ ] `.env` files are in `.gitignore`
- [ ] No `.env` files are committed to the repository
- [ ] Old hardcoded credentials are no longer valid
- [ ] All services start successfully with new credentials
- [ ] CI/CD pipelines use secure secret management
- [ ] Team members are informed of the change
- [ ] Documentation is updated

## Security Best Practices Going Forward

1. **Never commit credentials** to version control
2. **Use environment variables** for all configuration
3. **Rotate credentials regularly** (at least quarterly)
4. **Use strong, unique passwords** for each service
5. **Enable multi-factor authentication** where possible
6. **Monitor for suspicious access** to your services
7. **Use principle of least privilege** for service accounts
8. **Audit access logs** regularly

## Questions or Concerns?

If you have questions about this security migration or need assistance:

1. Review the [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md) documentation
2. Open an issue on the GitHub repository
3. Contact your security team if credentials may have been compromised

## Timeline

- **Immediate (Day 0):** Rotate all potentially exposed credentials
- **Within 24 hours:** Update all deployments with new environment-based configuration
- **Within 1 week:** Verify all services are running with new credentials
- **Ongoing:** Follow security best practices and regular credential rotation schedule

## Additional Resources

- [ENVIRONMENT_SETUP.md](./ENVIRONMENT_SETUP.md) - Complete environment variables guide
- [.env.example](./.env.example) - Template for root-level services
- [Backend/automation-engine/.env.example](./Backend/automation-engine/.env.example) - Template for automation engine

---

**Remember:** Security is everyone's responsibility. If you notice any security issues, report them immediately.
