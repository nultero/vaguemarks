use actix_web::{
    post, App, HttpResponse, HttpServer, Responder
};

use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize)]
struct Fruit {
    fruit_type: String,
    price: String,
}

#[post("/")]
async fn xml_to_json(req_body: String) -> impl Responder {

    #[allow(non_snake_case)] // strawberry is a single word dammit
    let mut Strawberry = Fruit{
        fruit_type: String::new(),
        price: String::new(),
    };

    'bodyparse : loop {
        let lines = req_body.lines();
        let mut seen_fruit = false;
        for line in lines {
            if line.contains("Strawberry") {
                let mut start_idx = 0;
                for (idx, c) in line.chars().enumerate() {
                    if c == 'S' { start_idx = idx }
                    else if c == '<' && start_idx > 0 {
                        Strawberry.fruit_type = line[start_idx..idx].to_owned();
                        seen_fruit = true;
                        break;
                    }
                }
            } else if seen_fruit {
                let mut start_idx = 0;
                for (idx, c) in line.chars().enumerate() {
                    if c == '>' {
                        start_idx = idx;
                    } else if c == '<' && start_idx > 0 {
                        Strawberry.price = line[start_idx..idx].to_owned();
                        break 'bodyparse;
                    }
                }
            }
        }
    }

    let s = serde_json::to_string(&Strawberry).unwrap();
    return HttpResponse::Ok().body(s);
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .service(xml_to_json)
    })
    .bind(("127.0.0.1", 3030))?
    .run()
    .await
}