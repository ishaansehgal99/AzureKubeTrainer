import torch
from torch.utils.data.distributed import DistributedSampler
from torch.utils.data import DataLoader
import torch.nn as nn
import torch.optim as optim

import torchvision
import torchvision.transforms as transforms

def evaluate(model, device, test_loader):
    model.eval()
    correct = 0
    total = 0
    with torch.no_grad():
        for data in test_loader:
            images, labels = data[0].to(device), data[1].to(device)
            outputs = model(images)
            _, predicted = torch.max(outputs.data, 1)
            total += labels.size(0)
            correct += (predicted == labels).sum().item()
    return correct / total

def main():
    torch.distributed.init_process_group(backend="nccl")

    # Setting up model and DDP
    model = torchvision.models.resnet18(pretrained=False)
    device = torch.device("cuda")
    model = model.to(device)
    model = torch.nn.parallel.DistributedDataParallel(model)

    # Data Preparations
    transform = transforms.Compose([
        transforms.RandomCrop(32, padding=4),
        transforms.RandomHorizontalFlip(),
        transforms.ToTensor(),
        transforms.Normalize((0.4914, 0.4822, 0.4465), (0.2023, 0.1994, 0.2010)),
    ])

    train_set = torchvision.datasets.CIFAR10(root="./data", train=True, download=True, transform=transform)
    test_set = torchvision.datasets.CIFAR10(root="./data", train=False, download=True, transform=transform)

    train_sampler = DistributedSampler(dataset=train_set)
    train_loader = DataLoader(dataset=train_set, batch_size=256, sampler=train_sampler, num_workers=2)
    test_loader = DataLoader(dataset=test_set, batch_size=128, shuffle=False, num_workers=2)

    criterion = nn.CrossEntropyLoss()
    optimizer = optim.SGD(model.parameters(), lr=0.1, momentum=0.9, weight_decay=1e-5)

    for epoch in range(10):
        model.train()
        for data in train_loader:
            inputs, labels = data[0].to(device), data[1].to(device)
            optimizer.zero_grad()
            outputs = model(inputs)
            loss = criterion(outputs, labels)
            loss.backward()
            optimizer.step()

        # Evaluating every epoch
        accuracy = evaluate(model=model, device=device, test_loader=test_loader)
        print(f"Epoch: {epoch}, Accuracy: {accuracy}")

    torch.save(model.state_dict(), "/dev/azureblob/model_weights.pth")

if __name__ == "__main__":
    main()
