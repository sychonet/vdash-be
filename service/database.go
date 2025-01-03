package service

import (
	"context"
	"log/slog"

	"github.com/sychonet/vdash-be/db"
	"github.com/sychonet/vdash-be/db/entity"
	"go.mongodb.org/mongo-driver/bson"
)

type DatabaseService struct {
	Host              string
	Port              string
	Username          string
	Password          string
	Name              string
	ServersCollection string
	IPsCollection     string
}

func NewDatabaseService(host, port, username, password, name, serversCollection, ipsCollection string) *DatabaseService {
	return &DatabaseService{
		Host:              host,
		Port:              port,
		Username:          username,
		Password:          password,
		Name:              name,
		ServersCollection: serversCollection,
		IPsCollection:     ipsCollection,
	}
}

func (d *DatabaseService) GetURI() string {
	return "mongodb://" + d.Username + ":" + d.Password + "@" + d.Host + ":" + d.Port
}

func (d *DatabaseService) AddServer(server entity.ServerInfo) error {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Insert the server information in database
	_, err := collection.InsertOne(context.Background(), server)
	if err != nil {
		slog.Error(err.Error())
	}

	return err
}

func (d *DatabaseService) AddServers(servers []interface{}) error {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Insert the server information in database
	_, err := collection.InsertMany(context.Background(), servers)
	if err != nil {
		slog.Error(err.Error())
	}

	return err
}

func (d *DatabaseService) CountServers() (int64, error) {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Count number of servers in database
	count, err := collection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		slog.Error(err.Error())
	}

	return count, err
}

func (d *DatabaseService) GetServers() ([]entity.ServerInfo, error) {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Get all servers from the database
	cursor, err := collection.Find(context.Background(), nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the servers
	var servers []entity.ServerInfo
	if err := cursor.All(context.Background(), &servers); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return servers, nil
}

func (d *DatabaseService) GetServer(id int) (*entity.ServerInfo, error) {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Get the server from the database
	var server entity.ServerInfo
	if err := collection.FindOne(context.Background(), entity.ServerInfo{ID: id}).Decode(&server); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return &server, nil
}

func (d *DatabaseService) DeleteServer(id int) error {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Delete the server from the database
	_, err := collection.DeleteOne(context.Background(), entity.ServerInfo{ID: id})
	if err != nil {
		slog.Error(err.Error())
	}

	return err
}

func (d *DatabaseService) AddIP(ip entity.IPInfo) error {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.IPsCollection)

	// Insert the IP information in database
	_, err := collection.InsertOne(context.Background(), ip)
	if err != nil {
		slog.Error(err.Error())
	}

	return err
}

func (d *DatabaseService) GetAvailablePublicIPs() ([]entity.IPInfo, error) {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.IPsCollection)

	// Get all available public IPs from the database
	cursor, err := collection.Find(context.Background(), bson.D{{Key: "available", Value: true}})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the IPs
	var ips []entity.IPInfo
	if err := cursor.All(context.Background(), &ips); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return ips, nil
}

func (d *DatabaseService) DeletePublicIP(ip string) error {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.IPsCollection)

	// Delete the IP from the database
	_, err := collection.DeleteOne(context.Background(), entity.IPInfo{IP: ip})
	if err != nil {
		slog.Error(err.Error())
	}

	return err
}

func (d *DatabaseService) UpdatePublicIP(ip string, serverID int, available bool) error {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.IPsCollection)

	// Update the IP information in database
	_, err := collection.UpdateOne(context.Background(), entity.IPInfo{IP: ip}, bson.D{{Key: "$set", Value: bson.D{{Key: "serverID", Value: serverID}, {Key: "available", Value: available}}}})
	if err != nil {
		slog.Error(err.Error())
	}

	return err
}

func (d *DatabaseService) CheckPublicIPAvailable(serverID int) (string, error) {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.IPsCollection)

	// Get the public IP associated with the server from the database
	var ip entity.IPInfo
	if err := collection.FindOne(context.Background(), entity.IPInfo{ServerID: serverID, Available: true}).Decode(&ip); err != nil {
		slog.Error(err.Error())
		return "", err
	}

	return ip.IP, nil
}

func (d *DatabaseService) GetServersWithIDs(serverIDs []int) ([]entity.ServerInfo, error) {
	// Connect to the database
	client := db.GetClient(d.GetURI())
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			slog.Error(err.Error())
		}
	}()

	// Get the collection
	collection := client.Database(d.Name).Collection(d.ServersCollection)

	// Get the servers from the database
	cursor, err := collection.Find(context.Background(), bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: serverIDs}}}})
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the servers
	var servers []entity.ServerInfo
	if err := cursor.All(context.Background(), &servers); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return servers, nil
}
