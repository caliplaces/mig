FROM php:8.1-apache

RUN apt-get update && apt-get install -y libpng-dev \
    && docker-php-ext-install gd

RUN rm -f /etc/apache2/mods-enabled/mpm_*.conf /etc/apache2/mods-enabled/mpm_*.load \
 && a2enmod mpm_prefork

RUN a2enmod rewrite

WORKDIR /var/www/html

COPY . .

RUN chown -R www-data:www-data /var/www/html

EXPOSE 80

CMD ["apache2-foreground"]

