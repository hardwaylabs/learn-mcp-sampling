#!/usr/bin/env python3
"""
Advanced Machine Learning Pipeline
A complete example showing data preprocessing, model training, and evaluation
"""

import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split, GridSearchCV
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.metrics import classification_report, confusion_matrix
import matplotlib.pyplot as plt
import seaborn as sns


class MLPipeline:
    """
    Machine Learning Pipeline for classification tasks
    """
    
    def __init__(self, random_state=42):
        self.random_state = random_state
        self.scaler = StandardScaler()
        self.label_encoder = LabelEncoder()
        self.model = None
        self.best_params = None
        
    def load_and_preprocess_data(self, filepath):
        """Load and preprocess the dataset"""
        df = pd.read_csv(filepath)
        
        # Handle missing values
        df = df.dropna()
        
        # Separate features and target
        X = df.drop('target', axis=1)
        y = df['target']
        
        # Encode categorical variables
        categorical_cols = X.select_dtypes(include=['object']).columns
        for col in categorical_cols:
            X[col] = self.label_encoder.fit_transform(X[col])
        
        return X, y
    
    def train_model(self, X, y, model_type='random_forest'):
        """Train the model with hyperparameter tuning"""
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=0.2, random_state=self.random_state
        )
        
        # Scale features
        X_train_scaled = self.scaler.fit_transform(X_train)
        X_test_scaled = self.scaler.transform(X_test)
        
        # Define models and parameter grids
        if model_type == 'random_forest':
            model = RandomForestClassifier(random_state=self.random_state)
            param_grid = {
                'n_estimators': [100, 200, 300],
                'max_depth': [10, 20, None],
                'min_samples_split': [2, 5, 10]
            }
        else:  # gradient_boosting
            model = GradientBoostingClassifier(random_state=self.random_state)
            param_grid = {
                'n_estimators': [100, 200],
                'learning_rate': [0.05, 0.1, 0.2],
                'max_depth': [3, 5, 7]
            }
        
        # Perform grid search
        grid_search = GridSearchCV(
            model, param_grid, cv=5, scoring='f1_weighted'
        )
        grid_search.fit(X_train_scaled, y_train)
        
        self.model = grid_search.best_estimator_
        self.best_params = grid_search.best_params_
        
        # Evaluate on test set
        y_pred = self.model.predict(X_test_scaled)
        
        return {
            'train_score': grid_search.best_score_,
            'test_predictions': y_pred,
            'test_actual': y_test,
            'classification_report': classification_report(y_test, y_pred)
        }
    
    def visualize_results(self, y_test, y_pred):
        """Create visualizations for model performance"""
        plt.figure(figsize=(12, 5))
        
        # Confusion matrix
        plt.subplot(1, 2, 1)
        cm = confusion_matrix(y_test, y_pred)
        sns.heatmap(cm, annot=True, fmt='d', cmap='Blues')
        plt.title('Confusion Matrix')
        plt.xlabel('Predicted')
        plt.ylabel('Actual')
        
        # Feature importance (for tree-based models)
        if hasattr(self.model, 'feature_importances_'):
            plt.subplot(1, 2, 2)
            importances = self.model.feature_importances_
            indices = np.argsort(importances)[::-1][:10]
            
            plt.bar(range(len(indices)), importances[indices])
            plt.title('Top 10 Feature Importances')
            plt.xlabel('Feature Index')
            plt.ylabel('Importance')
        
        plt.tight_layout()
        plt.show()


def main():
    """Main execution function"""
    # Initialize pipeline
    pipeline = MLPipeline(random_state=42)
    
    # Example usage (you would replace with actual data path)
    print("ML Pipeline Example")
    print("==================")
    print("This pipeline demonstrates:")
    print("1. Data preprocessing with missing value handling")
    print("2. Feature scaling and encoding")
    print("3. Hyperparameter tuning with GridSearchCV")
    print("4. Model evaluation and visualization")
    print("5. Support for multiple algorithms")
    
    # You would uncomment and modify these lines with actual data:
    # X, y = pipeline.load_and_preprocess_data('your_dataset.csv')
    # results = pipeline.train_model(X, y, model_type='random_forest')
    # pipeline.visualize_results(results['test_actual'], results['test_predictions'])
    # print(results['classification_report'])


if __name__ == "__main__":
    main()