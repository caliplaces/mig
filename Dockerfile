FROM php:8.4-apache

RUN apt-get update && apt-get install -y libpng-dev \
    && docker-php-ext-install gd \
    && rm -rf /var/lib/apt/lists/*

RUN a2enmod rewrite \
 && sed -i '/<Directory \/var\/www\/>/,/<\/Directory>/ s/AllowOverride None/AllowOverride All/' /etc/apache2/apache2.conf

WORKDIR /var/www/html

COPY . .

RUN chown -R www-data:www-data /var/www/html

EXPOSE 80

CMD ["apache2-foreground"]

