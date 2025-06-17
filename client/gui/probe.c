#include <gtk/gtk.h>
#include "probe.h"


void on_activate(GApplication *app) {
  g_print("Test\n");
  GtkWidget* window = gtk_application_window_new(GTK_APPLICATION(app));
  gtk_widget_show(window);
  on_create_window(window);
}

void new_application() {
  GtkApplication* app = gtk_application_new("com.github.wolmeister.wowa", G_APPLICATION_NON_UNIQUE);
  g_signal_connect_data(
    G_APPLICATION(app),
    "activate",
    G_CALLBACK(on_activate),
    NULL,
    NULL,
    0
  );
  g_application_run(G_APPLICATION(app), 0, NULL);
}