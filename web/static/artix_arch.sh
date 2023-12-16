#!/bin/sh
#
# FRESH INSTALLATION SCRIPT(Arch(systemd), Artix(openrc))
# Author: Michal < michal@michalkukla.xyz >
# Inspired and code from project: https://github.com/LukeSmithxyz/LARBS
# License: GNU GPLv3

DEVEL_DIR="$HOME/.local/src"
USER=$(whoami)

installpkg(){ pacman --noconfirm --needed -S "$1" ;}

die(){
	echo "$@" >&2
	exit 1
}

strap(){
	$(uname -r | grep -q "artix") && \
		basestrap /mnt base base-devel openrc elogind-openrc linux linux-firmware neovim rsync curl
	$(uname -r | grep -q "arch") && \
		pacstrap /mnt base base-devel linux linux-firmware neovim rsync curl
}
fstabgen_print(){
	case "$(readlink -f /sbin/init)" in
		*systemd*)
			genfstab -U /mnt
			;;
		*openrc*)
			fstabgen -U /mnt
			;;
	esac
}
fstabgen(){
# 	manually - cannot duplicate fd fstabgen: too many open files
	
	case "$(readlink -f /sbin/init)" in
		*systemd*)
			echo "genfstab -U /mnt > /mnt/etc/fstab"
			;;
		*openrc*)
			echo "Manually: fstabgen -U /mnt > /mnt/etc/fstab"
			;;
	esac
}

changeroot(){
	uname -r | grep -q "artix" && artix-chroot /mnt ||
		{ uname -r | grep -q "arch" && arch-chroot /mnt; }
# 	artix-chroot /mnt
#	arch-chroot /mnt
}

rootpasswd(){
	passwd
}
systembeepoff(){
	rmmod pcspkr
	echo "blacklist pcspkr" > /etc/modprobe.d/nobeep.conf
}

hostname(){
	hostname=$1
	[ -z "$hostname" ] && die "Usage: $PROGRAM $COMMAND hostname"
	echo "$hostname" > /etc/hostname
	
}
simplehostsfile(){
	hostname=$1
	[ -z "$hostname" ] && die "Usage: $PROGRAM $COMMAND hostname"
	echo "127.0.0.1		localhost" >> /etc/hosts
	echo "::1		localhost" >> /etc/hosts
	echo "127.0.0.1		$hostname.localdomain $hostname" >> /etc/hosts
}

set_zone(){
	hwclock --systohc
	ZONE="$1"
	#ln -svf /usr/share/zoneinfo/Europe/Prague /etc/localtime
	[ -z "$ZONE" ] && ZONE="Europe/Prague" # die "Usage: $PROGRAM $COMMAND zone"
	ln -svf /usr/share/zoneinfo/"$ZONE" /etc/localtime
}
intel_graphics(){
	installpkg "xf86-video-intel"
	sed -i "s/^MODULES=()$/MODULES=(i915)/" /etc/mkinitcpio.conf
	mkinitcpio -p linux

	echo "options i915 enable_guc=2" > /etc/modprobe.d/i915.conf
	echo "Section \"Device\"
                Identifier \"Intel Graphics\"
                Driver \"intel\"
                Option \"TearFree\" \"true\"
                Option \"DRI\"    \"3\"
	EndSection" > /etc/X11/xorg.conf.d/20-intel.conf
}

amd_graphics(){
	die "Not implemented yet"
}

charset(){
	echo "LANG=en_US.UTF-8" >> /etc/locale.conf
	echo "LC_COLLATE=cs_CZ.UTF-8" >> /etc/locale.conf
	
	echo "cs_CZ ISO-8859-2" >> /etc/locale.gen
	echo "cs_CZ.UTF-8 UTF-8" >> /etc/locale.gen
	echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen
	echo "en_US ISO-8859-1" >> /etc/locale.gen

	locale-gen
}

init_networkmanager(){
	installpkg "networkmanager"
	case "$(readlink -f /sbin/init)" in
		*systemd*)
			systemctl enable NetworkManager
			systemctl start NetworkManager
			;;
		*openrc*)
			installpkg "networkmanager-openrc"
			rc-update add NetworkManager default
			rc-service NetworkManager start
			;;
	esac

	nmcli connection modify Wired\ connection\ 1 +connection.id "Home" +ipv4.method "manual" +ipv4.addresses "10.120.29.73/24" \
	+ipv4.gateway "10.120.29.1" +ipv4.routes "192.168.1.0/24 10.120.29.79" +ipv4.dns "10.120.0.250,10.120.0.251"

#	nmcli connection down Home
#	sleep 1
#	nmcli connection up Home

}

create_dirs_links(){
	# nearly work - update_list_pkg.hook not found
	name=$1
	[ -z "$name" ] && die "Usage: $PROGRAM $COMMAND name"
	if [ "$USER" = "root" ]; then
		mkdir -pv /media/$name/{HardDrive,ExtDrive,FlashDrive}
	else
		mkdir -pv $DEVEL_DIR
		mkdir -pv $HOME/phone
		ln -sv /media/$name/ExtDrive/ $HOME/extdrive
		ln -sv /media/$name/FlashDrive/ $HOME/flashdrive
		ln -sv /media/$name/HardDrive/ $HOME/data 
		ln -sv /home/$name/.config/zsh/.zprofile /home/$name/.zprofile
		sudo mkdir -pv /etc/pacman.d/hooks 
		sudo ln -sv /home/$name/.config/paccman/hooks/update_list_pkg.hook /etc/pacman.d/hooks/update_list_pkg.hook
	fi
	# check if hard drive is mounted
#	mount | grep "/media/$name/HardDrive" && \
#		{ln -s /media/$name/HardDrive/Movies $HOME/movies;
#	ln -s /media/$name/HardDrive/Music $HOME/music }
}

create_user(){
	# work
	name="$1"
	[[ -z $name ]] && die "Usage: $PROGRAM $COMMAND username"
	useradd -m -g wheel -s /bin/bash "$name" ||
		usermod -a -G wheel "$name" && mkdir -p /home/"$name" && chown "$name":wheel /home/"$name"
	passwd $name
}
sudo_setup(){
	# work
	# sudoers %wheel ALL=(ALL) ALL
	echo "%wheel ALL=(ALL:ALL) ALL" >/etc/sudoers.d/00-wheel
	echo "Defaults \!tty_tickets" >>/etc/sudoers.d/00-wheel
}
laptop_touchpad(){
	# Enable tap to click
	[ ! -f /etc/X11/xorg.conf.d/40-libinput.conf ] && printf 'Section "InputClass"
		Identifier "libinput touchpad catchall"
		MatchIsTouchpad "on"
		MatchDevicePath "/dev/input/event*"
		Driver "libinput"
		# Enable left mouse button by tapping
		Option "Tapping" "on"
	EndSection' >/etc/X11/xorg.conf.d/40-libinput.conf
}

makepkg_cores(){
	# Use all cores for compilation.
	sed -i "s/-j2/-j$(nproc)/;s/^#MAKEFLAGS/MAKEFLAGS/" /etc/makepkg.conf
}

yay_bin(){
	[[ $(command -v git) ]] || sudo pacman --noconfirm --needed -S "git"
	git clone https://aur.archlinux.org/yay.git $DEVEL_DIR/yay
	cd $DEVEL_DIR/yay
	makepkg -si
	cd $HOME
}


pacman_setup(){
	# nearly work, forgot resolv.conf
	# Make pacman colorful, concurrent downloads and Pacman eye-candy.
	grep -q "ILoveCandy" /etc/pacman.conf || sed -i "/#VerbosePkgLists/a ILoveCandy" /etc/pacman.conf
	sed -i "s/^#ParallelDownloads = 5$/ParallelDownloads = 5/;s/^#Color$/Color/" /etc/pacman.conf

	case "$(readlink -f /sbin/init)" in
		*systemd* )
			pacman --noconfirm -S archlinux-keyring ;;
		*)
			if ! grep -q "^\[universe\]" /etc/pacman.conf; then
				echo "[universe]
	Server = https://universe.artixlinux.org/\$arch
	Server = https://mirror1.artixlinux.org/universe/\$arch
	Server = https://mirror.pascalpuffke.de/artix-universe/\$arch
	Server = https://artixlinux.qontinuum.space/artixlinux/universe/os/\$arch
	Server = https://mirror1.cl.netactuate.com/artix/universe/\$arch
	Server = https://ftp.crifo.org/artix-universe/" >>/etc/pacman.conf
				pacman -Sy --noconfirm
			fi
			pacman --noconfirm --needed -S artix-keyring artix-archlinux-support
			for repo in extra community; do
				grep -q "^\[$repo\]" /etc/pacman.conf ||
					echo "[$repo]
Include = /etc/pacman.d/mirrorlist-arch" >> /etc/pacman.conf
			done
			pacman -Sy
			pacman-key --populate archlinux
	esac
}

install_pacman_pkgs(){
        user=$1
        [[ -z $name ]] && die "Usage: $PROGRAM $COMMAND user"
        for n in $(cat /home/$user/.config/paccman/pkglist_$(cat /etc/hostname).txt); do
                pacman -S --noconfirm --needed $n
        done
#       HOST=$(cat /etc/hostname)
#       sudo pacman -S --needed < $HOME/.config/paccman/pkglist_$HOST.txt
}
install_aur_pkgs(){
	HOST=$(cat /etc/hostname)
	for n in $(cat ~/.config/paccman/pkglist_$HOST.txt); do
		yay -S "$n"
	done
	
}
install_pip_pkgs(){
	for n in tldextract; do
		pip3 install $n
	done
}

git_clone(){
	[[ -d $DEVEL_DIR ]] || mkdir -p $HOME/devel
	[[ $(command -v git) ]] || installpkg "git"

	for n in $(cat $HOME/scripts/fresh_install/clone.txt); do
		cd $DEVEL_DIR"/"
		git clone $n
		n=$(echo "${n#*/}")
		n=$(echo "${n%.git}")

		cd $DEVEL_DIR"/"$n || exit 1
		make 
		sudo make install 
	done
	mv -v $DEVEL_DIR/fork-dwm/ $DEVEL_DIR/dwm/
	mv -v $DEVEL_DIR/fork-st/ $DEVEL_DIR/st/
	mv -v $DEVEL_DIR/fork-dmenu/ $DEVEL_DIR/dmenu/
	mv -v $DEVEL_DIR/fork-st-terminal/ $DEVEL_DIR/st/
	mv -v $DEVEL_DIR/fork-password-store/ $DEVEL_DIR/password-store/
}
make_gitrepo(){
	for n in dwm st dmenu password-store dwmblocks zaread rsync-script xkblayout-state; do
		cd $DEVEL_DIR/$n
		sudo make install
	done
}
boot_loader(){
#	manually: grub-install - wrong device
	[[ $(command -v grub-install) ]] || installpkg "grub"
	if [[ -f /sys/firmware/efi/fw_platform_size ]]; then
# 		UEFI
		installpkg "efibootmgr"
		grub-install --target=x86_64-efi --efi-directory=/boot --bootloader-id=GRUB
		grub-mkconfig -o /boot/grub/grub.cfg
	else
#		BIOS
		echo "Check for yourself the correct device /dev/sdXx"
		echo "Manually: grub-install --target=i386-pc /dev/sdXx"
		echo "		grub-mkconfig -o /boot/grub/grub.cfg"
		

	fi
	echo "optinally:"
	echo "sed -i "s/^GRUB_TIMEOUT=[0-9]*$/GRUB_TIMEOUT=10/" /etc/default/grub"
}

vimplugininstall(){
	#nearly work, command nvim -c not working
	#Minimalist plugin manager for vim
	sh -c 'curl -fLo "${XDG_DATA_HOME:-$HOME/.local/share}"/nvim/site/autoload/plug.vim --create-dirs \
	       https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim' 
	nvim -c "PlugInstall|q|q"
}


sshd(){
	[[ $(command -v ssh) ]] || installpkg "openssh"

	sed -i "s/^#AllowTcpForwading .*/AllowTcpForwarding yes/;
		s/^#X11Forwarding .*/X11Forwarding yes/;
		s/^#X11DisplayOffset .*/X11DisplayOffset 10/;
		s/^#X11UseLocalhost .*/X11UseLocalhost yes/" /etc/ssh/sshd_config

	case "$(readlink -f /sbin/init)" in
		*systemd* ) 
			systemctl enable sshd.service
			systemctl start sshd.service
			;;	
		*openrc* )
			installpkg "openssh-openrc" 
#			rc-update add sshd default
#			rc-service sshd start
			;;
	esac
}

importgpg_key(){
	mkdir -pv $XDG_DATA_HOME/gnupg && chmod -v go-rx $XDG_DATA_HOME/gnupg
	ln -s $XDG_CONFIG_HOME/gnupg/gpg-agent.conf $XDG_DATA_HOME/gnupg/gpg-agent.conf
	gpg --import my-key1.asc
	gpg --edit-key  # trust 5
}

set_hosts(){
#	git clone git@github.com:Solamil/hosts.git $DEVEL_DIR/hosts
#	rs hosts/ pull -r -n
	cd $DEVEL_DIR/hosts/
	git checkout master
	git remote add upstream git@github.com:StevenBlack/hosts.git
	git remote -v
	git fetch upstream
	git merge
	echo "Fix conflicts manually"
}

init_mbsync(){
	XDG_DATA_HOME=${XDG_DATA_HOME:-$HOME"/.local/share"}
	XDG_CONFIG_HOME=${XDG_CONFIG_HOME:-$HOME"/.config"}
	# init mbsync download mails 1.) Sync All -> Sync Pull, Create Both -> Create Near
	for i in $( grep "^Path" $XDG_CONFIG_HOME/mbsync/config | grep -o "[^ ]*$" ); do
		maildir=$(dirname -- $i)	
		mkdir -p -v $maildir/{mail,cache/{bodies,headers}} 
	done
	sed -i "s/^Sync All/Sync Pull/g;s/^Create Both/Create Near/g" $XDG_CONFIG_HOME/mbsync/config
#	mbsync -c $XDG_CONFIG_HOME/mbsync/config -a && {
#		# 2.) Sync Pull -> Sync All, Create Near -> Create Both
#		sed -i "s/^Sync Pull/Sync All/g;s/^Create Near/Create Both/g" $XDG_CONFIG_HOME/mbsync/config
#	} &
	mbsync -c $XDG_CONFIG_HOME/mbsync/config -a
	# 2.) Sync Pull -> Sync All, Create Near -> Create Both
	sed -i "s/^Sync Pull/Sync All/g;s/^Create Near/Create Both/g" $XDG_CONFIG_HOME/mbsync/config

}

init_cups(){
	# CUPS
#	yay -S samsung-unified-driver

	case "$(readlink -f /sbin/init)" in
		*systemd*)
			systemctl enable cups.service 
			systemctl start cups.service 
			;;
		*openrc*)
			# installpkg "cupsd-openrc"
			rc-update add cupsd default
			rc-service cupsd start
			;;
	esac

	MODEL_PRINTER="M2070"
	NAME_PRINTER="SAMSUNG_PRINTER"
	# lpinfo -v parse "usb:..."
	
	ppdgz_file=$(lpinfo -m | grep "$MODEL_PRINTER")

	if [[ -n $ppdgz_file ]]; then
		lpadmin -p $NAME_PRINTER -E -v "usb://Samsung/M2070%20Series?serial=ZF44B8KH3C01MYN&interface=1" -m "$ppdgz_file"
		lpoptions -d $NAME_PRINTER
	else
		die "ppd.gz file for $MODEL_PRINTER was not found. Try find it manually."
	fi

}

autologin(){

	case "$(readlink -f /sbin/init)" in
		*systemd*)
			# arch wiki agetty
			echo "This content to the getty@tty1.service:
			[Service]
			ExecStart=
			ExecStart=-/sbin/agetty -o '-p -f -- \\u' --noclear --autologin username - \$TERM"
			systemctl edit getty@tty1.service
			;;
		*openrc*)
			autologin_option="agetty_options=\"--autologin $USER --noclear\""
			sudo cp /etc/init.d/agetty.tty1 /etc/init.d/agetty-autologin.tty1
			commandargs_line=$(grep "command_args_foreground=.*" /etc/init.d/agetty-autologin.tty1)
			sudo sed -i "s/$commandargs_line/$autologin_option\n$commandargs_line/" /etc/init.d/agetty-autologin.tty1
			
			sudo rc-update del agetty.tty1
			sudo rc-update add agetty-autologin.tty1 default
			;;
	esac

}

hibernation(){
	# Make hibernation work
	# HOOKS append resume
	sudo sed -i "s/filesystems keyboard/filesystems resume keyboard/" /etc/mkinitcpio.conf
	sudo mkinitcpio -p linux

	uuid=$(grep "swap" /etc/fstab | grep -o "^UUID=[a-f0-9-]*")
	grub_line=$( grep -o "^GRUB_CMDLINE_LINUX_DEFAULT=\"[^\"]*" /etc/default/grub )
	# GRUB_CMDLINE_DEFAULT = resume=UUID=...
	sudo sed -i "s/$grub_line/$grub_line resume=$uuid/" /etc/default/grub
	sudo grub-mkconfig -o /boot/grub/grub.cfg

}
set_default_shell(){
	name=$(whoami)
	chsh -s /bin/zsh "$name"
	sudo -u "$name" mkdir -p "/home/$name/.local/cache/zsh/"
}

set_remap_mic_to_mono(){
	# Lenovo ntb
	[ "$(cat /etc/hostname)" = "lenovo" ] && printf '### Remap microphone to mono
load-module module-remap-source source_name=record_mono master=alsa_input.pci-0000_00_1f.3.analog-stereo master_channel_map=front-left channel_map=mono
set-default-source record_mono' >> /etc/pulse/default.pa
}

set_correct_time(){
	case "$(readlink -f /sbin/init)" in
		*systemd*)
			timedatectl set-ntp true ;;
		*openrc*)
			installpkg "ntp-openrc"
			rc-update add ntp-client default
			rc-service ntp-client start
			;;
	esac
}

append_hdd_drive(){
	name=$1
	[[ -z $name ]] && die "Usage: $PROGRAM $COMMAND name"
	[[ $(cat /etc/hostname) == "desktop" ]] && UUID_HDD="UUID=\"48FE2527FE250EAE\""
	[[ $(cat /etc/hostname) == "lenovo" ]] && UUID_HDD="UUID=\"D8D82351D8232D66\""

	HDD_MOUNT_POINT="/media/$name/HardDrive"
	[[ -d $HDD_MOUNT_POINT ]] || mkdir -pv $HDD_MOUNT_POINT
	echo "$UUID_HDD" "$HDD_MOUNT_POINT"
	echo "$UUID_HDD $HDD_MOUNT_POINT ntfs defaults,umask=022,fmask=133,uid=1000,gid=998 0 0" >> /etc/fstab

}

PROGRAM="${0##*/}"
COMMAND="$1"

case "$1" in
	strap) strap ;;
	fstabgen) fstabgen ;;
	chroot) changeroot ;;
	beepoff) systembeepoff ;;
	hostname) shift; hostname "$1" ;;
	hosts) shift; simplehostsfile "$1" ;;
	charset) charset ;;
	timezone) shift; set_zone "$@" ;;
	makepkg-cores) makepkg_cores ;;
	setupacman) pacman_setup ;;
	networkmanager) init_networkmanager ;;
	bootloader) boot_loader ;;
	createuser) shift; create_user "$@" ;;
	vimplug) vimplugininstall ;;
	dir-links) shift; create_dirs_links "$@" ;;
	default-shell) set_default_shell ;;
	sudosetup) sudo_setup ;;
	install-yay) yay_bin ;;	
	pacmanpkgs) install_pacman_pkgs ;;
	aurpkgs) install_aur_pkgs ;;
	pypkgs) install_pip_pkgs ;;
	git-repos) git_clone ;;
	makegitrepo) make_gitrepo ;;
	openssh-setup) sshd ;;
	printer) init_cups ;;
	hdd-fstab) append_hdd_drive ;;
	mbsync) init_mbsync ;;
	intel-graphics) intel_graphics ;;
	amd-graphics) amd_graphics ;;
	gpgkey) importgpg_key ;;
	autologin) autologin ;;
	hibernation) hibernation ;;
	stevenhosts) set_hosts ;;
	correct-time) set_correct_time ;;
#	dir-links) create_dirs_links ;;
	touchpad) laptop_touchpad ;;
	micremap) set_remap_mic_to_mono ;;
	installpkg) shift; installpkg "$1" ;;
	*) die "Wrong command: $1" ;;
esac

wait
exit 0
