#! /usr/bin/env python3


from PIL import Image, ImageDraw
import face_recognition
import pika
import base64
import io
from numpy import array


BASE64_PNG_PREFIX = b'data:image/png;base64,'
BASE64_PNG_PREFIX_LEN = len(BASE64_PNG_PREFIX)


def detect_face(image):
    face_locations = face_recognition.face_locations(image)
    # print(face_locations)

    # for face_location in face_locations:
    #     top, right, bottom, left = face_location
    #     print("A face is located at pixel location Top: {}, Left: {}, Bottom: {}, Right: {}".format(top, left, bottom,
    #                                                                                                 right))
    #     face_image = image[top:bottom, left:right]
    #     pil_image = Image.fromarray(face_image)
    #     pil_image.show()

    face_landmarks_list = face_recognition.face_landmarks(image)
    # print(face_landmarks_list)

    for face_landmarks in face_landmarks_list:

        # Print the location of each facial feature in this image
        facial_features = [
            'chin',
            'left_eyebrow',
            'right_eyebrow',
            'nose_bridge',
            'nose_tip',
            'left_eye',
            'right_eye',
            'top_lip',
            'bottom_lip'
        ]

        # for facial_feature in facial_features:
        #     print("The {} in this face has the following points: {}".format(facial_feature,
        #                                                                     face_landmarks[facial_feature]))

        # Let's trace out each facial feature in the image with a line!
        pil_image = Image.fromarray(image)
        d = ImageDraw.Draw(pil_image)

        for facial_feature in facial_features:
            d.line(face_landmarks[facial_feature], width=3)

        # pil_image.show()
        pil_image.save('test5.png', format='PNG')

        png_img_buffer = io.BytesIO()
        pil_image.save(png_img_buffer, format='PNG')
        img_str_base64 = base64.b64encode(png_img_buffer.getvalue())

        return BASE64_PNG_PREFIX + img_str_base64


def callback(ch, method, props, body):
    body = body[BASE64_PNG_PREFIX_LEN:]
    b = io.BytesIO(base64.b64decode(body))
    im = Image.open(b)
    im = im.convert('RGB')
    png_base64 = detect_face(array(im))

    if png_base64:
        properties = pika.BasicProperties(correlation_id=props.correlation_id)
        ch.basic_publish(exchange='', routing_key=props.reply_to, properties=properties, body=png_base64)
        print('Published message')

    ch.basic_ack(delivery_tag=method.delivery_tag)


if __name__ == '__main__':
    # image = face_recognition.load_image_file('test.png')
    # print(type(image))

    rpc_queue_name = 'rpc_queue'
    conn = pika.BlockingConnection(pika.ConnectionParameters(host='35.201.229.242'))
    channel = conn.channel()
    channel.queue_declare(queue=rpc_queue_name)
    channel.basic_consume(callback, queue=rpc_queue_name)

    print('[*] waiting for message. To exit, press CTRL+C')
    channel.start_consuming()
